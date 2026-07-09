package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/will-x86/bdns/dns/pkg/db/models"
)

// listCategories returns the distinct blocklist categories, for UI dropdowns.
func (a *API) listCategories(c fiber.Ctx) error {
	cats, err := a.repo.ListCategories(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list categories")
	}
	return c.JSON(cats)
}

// ---- permanent whitelist ----

func (a *API) listPermanentWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	out, err := a.repo.ListPermanentWhitelist(c.Context(), pid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list whitelist")
	}
	return c.JSON(out)
}

func (a *API) addPermanentWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	domain, err := parseDomainBody(c)
	if err != nil {
		return err
	}
	if err := a.repo.AddPermanentWhitelist(c.Context(), pid, domain); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add domain")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deletePermanentWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeletePermanentWhitelist(c.Context(), pid, c.Params("domain")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete domain")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- temporary whitelist ----

func (a *API) listTemporaryWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	out, err := a.repo.ListTemporaryWhitelist(c.Context(), pid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list whitelist")
	}
	return c.JSON(out)
}

func (a *API) addTemporaryWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	var req tempWhitelistReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Domain = strings.TrimSpace(req.Domain)
	if req.Domain == "" {
		return fiber.NewError(fiber.StatusBadRequest, "domain is required")
	}
	expiresAt := req.ExpiresAt
	if expiresAt == 0 && req.TTLSeconds > 0 {
		expiresAt = time.Now().Unix() + req.TTLSeconds
	}
	if expiresAt <= time.Now().Unix() {
		return fiber.NewError(fiber.StatusBadRequest, "expires_at must be in the future (or provide ttl_seconds)")
	}
	if err := a.repo.AddTemporaryWhitelist(c.Context(), pid, req.Domain, expiresAt); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add domain")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deleteTemporaryWhitelist(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeleteTemporaryWhitelist(c.Context(), pid, c.Params("domain")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete domain")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- category blocks ----

func (a *API) listCategoryBlocks(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	out, err := a.repo.ListCategoryBlocks(c.Context(), pid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list category blocks")
	}
	return c.JSON(out)
}

func (a *API) addCategoryBlock(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	var req categoryReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category is required")
	}
	if err := a.repo.AddCategoryBlock(c.Context(), pid, req.Category); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add category block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deleteCategoryBlock(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeleteCategoryBlock(c.Context(), pid, c.Params("category")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete category block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- time blocks ----

func (a *API) listTimeBlocks(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	out, err := a.repo.ListTimeBlocks(c.Context(), pid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list time blocks")
	}
	return c.JSON(out)
}

func (a *API) addTimeBlock(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	var req timeBlockReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category is required")
	}
	// 15-min slots 0..95, day 0..7 (DB CHECK constraints).
	if req.StartTime < 0 || req.StartTime > 95 || req.EndTime < 0 || req.EndTime > 95 {
		return fiber.NewError(fiber.StatusBadRequest, "start_time and end_time must be within 0..95")
	}
	if req.EndTime < req.StartTime {
		return fiber.NewError(fiber.StatusBadRequest, "end_time must be >= start_time")
	}
	if req.Day < 0 || req.Day > 7 {
		return fiber.NewError(fiber.StatusBadRequest, "day must be within 0..7")
	}
	tb := models.TimeBlock{
		ProfileID: pid,
		Category:  req.Category,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Day:       req.Day,
	}
	if err := a.repo.AddTimeBlock(c.Context(), tb); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add time block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// deleteTimeBlock removes the row matching ?category=&day=&start_time=&end_time=.
func (a *API) deleteTimeBlock(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	category := strings.TrimSpace(c.Query("category"))
	if category == "" {
		return fiber.NewError(fiber.StatusBadRequest, "category query param is required")
	}
	day, dErr := strconv.Atoi(c.Query("day"))
	start, sErr := strconv.Atoi(c.Query("start_time"))
	end, eErr := strconv.Atoi(c.Query("end_time"))
	if dErr != nil || sErr != nil || eErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "day, start_time and end_time query params must be integers")
	}
	if err := a.repo.DeleteTimeBlock(c.Context(), pid, category, day, start, end); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete time block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// parseDomainBody decodes a {"domain": "..."} body and validates it.
func parseDomainBody(c fiber.Ctx) (string, error) {
	var req domainReq
	if err := bind(c, &req); err != nil {
		return "", err
	}
	req.Domain = strings.TrimSpace(req.Domain)
	if req.Domain == "" {
		return "", fiber.NewError(fiber.StatusBadRequest, "domain is required")
	}
	return req.Domain, nil
}
