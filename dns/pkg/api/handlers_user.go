package api

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// createUser is the public bootstrap: mints a user and returns its token once.
func (a *API) createUser(c fiber.Ctx) error {
	var req createUserReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Timezone = strings.TrimSpace(req.Timezone)
	if req.Timezone == "" {
		return fiber.NewError(fiber.StatusBadRequest, "timezone is required")
	}

	token, err := newToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not generate token")
	}
	user, err := a.repo.CreateUser(c.Context(), req.Timezone, token)
	if err != nil {
		a.log.Error().Err(err).Msg("create user failed")
		return fiber.NewError(fiber.StatusInternalServerError, "could not create user")
	}
	return c.Status(fiber.StatusCreated).JSON(userWithToken{User: user, APIToken: token})
}

func (a *API) getMe(c fiber.Ctx) error {
	return c.JSON(currentUser(c))
}

func (a *API) updateMe(c fiber.Ctx) error {
	var req updateMeReq
	if err := bind(c, &req); err != nil {
		return err
	}
	req.Timezone = strings.TrimSpace(req.Timezone)
	if req.Timezone == "" {
		return fiber.NewError(fiber.StatusBadRequest, "timezone is required")
	}
	if err := a.repo.UpdateUserTimezone(c.Context(), currentUser(c).ID, req.Timezone); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update user")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deleteMe(c fiber.Ctx) error {
	if err := a.repo.DeleteUser(c.Context(), currentUser(c).ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete user")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// regenerateToken issues a fresh api_token, invalidating the old one.
func (a *API) regenerateToken(c fiber.Ctx) error {
	token, err := newToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not generate token")
	}
	if err := a.repo.SetUserToken(c.Context(), currentUser(c).ID, token); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update token")
	}
	return c.JSON(fiber.Map{"api_token": token})
}

// ---- friends ----

func (a *API) listFriends(c fiber.Ctx) error {
	friends, err := a.repo.ListFriends(c.Context(), currentUser(c).ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not list friends")
	}
	return c.JSON(friends)
}

func (a *API) addFriend(c fiber.Ctx) error {
	var req addFriendReq
	if err := bind(c, &req); err != nil {
		return err
	}
	me := currentUser(c).ID
	req.FriendID = strings.TrimSpace(req.FriendID)
	if req.FriendID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "friend_id is required")
	}
	if req.FriendID == me {
		return fiber.NewError(fiber.StatusBadRequest, "cannot friend yourself")
	}
	exists, err := a.repo.UserExists(c.Context(), req.FriendID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	if err := a.repo.AddFriend(c.Context(), me, req.FriendID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add friend")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *API) deleteFriend(c fiber.Ctx) error {
	if err := a.repo.DeleteFriend(c.Context(), currentUser(c).ID, c.Params("friendID")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove friend")
	}
	return c.SendStatus(fiber.StatusNoContent)
}
