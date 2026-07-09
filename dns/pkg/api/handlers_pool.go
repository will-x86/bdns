package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

func (a *API) listPools(c fiber.Ctx) error {
	pools, err := a.repo.ListPoolsForUser(c.Context(), currentUser(c).ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list pools")
	}
	return c.JSON(pools)
}

func (a *API) createPool(c fiber.Ctx) error {
	var req createPoolReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if req.PoolMode != "shared" && req.PoolMode != "borrow" {
		return fiber.NewError(fiber.StatusBadRequest, "pool_mode must be 'shared' or 'borrow'")
	}
	if req.TotalLimit <= 0 {
		req.TotalLimit = 6000 // matches the DB default
	}
	pool, err := a.repo.CreatePool(c.Context(), currentUser(c).ID, req.Name, req.PoolMode, req.TotalLimit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create pool")
	}
	return c.Status(fiber.StatusCreated).JSON(pool)
}

func (a *API) getPool(c fiber.Ctx) error {
	pool, err := a.accessiblePool(c)
	if err != nil {
		return err
	}
	return c.JSON(pool)
}

func (a *API) updatePool(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
	if err != nil {
		return err
	}
	var req updatePoolReq
	if err := bind(c, &req); err != nil {
		return err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = pool.Name
	}
	limit := req.TotalLimit
	if limit <= 0 {
		limit = pool.TotalLimit
	}
	if err := a.repo.UpdatePool(c.Context(), pool.ID, name, limit); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update pool")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deletePool(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeletePool(c.Context(), pool.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete pool")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- members ----

func (a *API) listPoolMembers(c fiber.Ctx) error {
	pool, err := a.accessiblePool(c)
	if err != nil {
		return err
	}
	members, err := a.repo.ListPoolMembers(c.Context(), pool.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list members")
	}
	return c.JSON(members)
}

// addPoolMember adds a profile owned by the creator or one of their friends.
func (a *API) addPoolMember(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
	if err != nil {
		return err
	}
	var req addMemberReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.ProfileID = strings.TrimSpace(req.ProfileID)
	if req.ProfileID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "profile_id is required")
	}

	profile, err := a.repo.GetProfile(c.Context(), req.ProfileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if profile == nil {
		return fiber.NewError(fiber.StatusNotFound, "profile not found")
	}
	creator := currentUser(c).ID
	if profile.UserID != creator {
		isFriend, err := a.repo.IsFriend(c.Context(), creator, profile.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "internal error")
		}
		if !isFriend {
			return fiber.NewError(fiber.StatusForbidden, "profile must belong to you or a friend")
		}
	}

	if err := a.repo.AddPoolMember(c.Context(), pool.ID, req.ProfileID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add member")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deletePoolMember(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeletePoolMember(c.Context(), pool.ID, c.Params("profileID")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove member")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- pool category blocks ----

func (a *API) listPoolCategoryBlocks(c fiber.Ctx) error {
	pool, err := a.accessiblePool(c)
	if err != nil {
		return err
	}
	blocks, err := a.repo.ListPoolCategoryBlocks(c.Context(), pool.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list category blocks")
	}
	return c.JSON(blocks)
}

func (a *API) addPoolCategoryBlock(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
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
	if err := a.repo.AddPoolCategoryBlock(c.Context(), pool.ID, req.Category); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add category block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deletePoolCategoryBlock(c fiber.Ctx) error {
	pool, err := a.managedPool(c)
	if err != nil {
		return err
	}
	if err := a.repo.DeletePoolCategoryBlock(c.Context(), pool.ID, c.Params("category")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete category block")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- live limits ----

// getPoolLimits reports live remaining from the pool store; nil where unknown.
func (a *API) getPoolLimits(c fiber.Ctx) error {
	pool, err := a.accessiblePool(c)
	if err != nil {
		return err
	}
	ctx := c.Context()
	resp := poolLimitsResp{PoolID: pool.ID, Mode: pool.PoolMode, TotalLimit: pool.TotalLimit}

	if pool.PoolMode == "shared" {
		if remaining, err := a.pool.GetRemainingShared(ctx, pool.ID); err == nil {
			resp.Remaining = &remaining
		}
		return c.JSON(resp)
	}

	members, err := a.repo.ListPoolMembers(ctx, pool.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list members")
	}
	resp.Members = make([]memberLimit, 0, len(members))
	for _, m := range members {
		ml := memberLimit{ProfileID: m.ProfileID}
		if remaining, err := a.pool.GetRemainingBorrow(ctx, pool.ID, m.ProfileID); err == nil {
			ml.Remaining = &remaining
		}
		resp.Members = append(resp.Members, ml)
	}
	return c.JSON(resp)
}
