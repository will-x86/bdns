package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func (a *API) listProfiles(c fiber.Ctx) error {
	profiles, err := a.repo.ListProfiles(c.Context(), currentUser(c).ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list profiles")
	}
	return c.JSON(profiles)
}

func (a *API) createProfile(c fiber.Ctx) error {
	var req createProfileReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	profile, err := a.repo.CreateProfile(c.Context(), currentUser(c).ID, req.Name)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create profile")
	}
	return c.Status(fiber.StatusCreated).JSON(profile)
}

func (a *API) getProfile(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	profile, err := a.repo.GetProfile(c.Context(), pid)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	return c.JSON(profile)
}

func (a *API) updateProfile(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	var req updateProfileReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if err := a.repo.UpdateProfileName(c.Context(), pid, req.Name); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update profile")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deleteProfile(c fiber.Ctx) error {
	pid, err := a.ownedProfile(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeleteProfile(c.Context(), pid); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete profile")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
