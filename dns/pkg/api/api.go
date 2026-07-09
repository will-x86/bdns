// Package api is the JSON management API (Fiber v3) for bdns config. See API.md.
package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/db/models"
	"github.com/will-x86/bdns/dns/pkg/store"
)

// API holds the dependencies shared by every handler.
type API struct {
	repo *db.Repo
	pool store.Pool // live per-pool query limits (valkey or memory)
	log  zerolog.Logger
}

const userLocal = "user"

// Serve starts the management API on addr and blocks until it stops or ctx is
// cancelled.
func Serve(ctx context.Context, addr string, repo *db.Repo, pool store.Pool) error {
	log := zerolog.Ctx(ctx).With().Str("component", "api").Logger()
	a := &API{repo: repo, pool: pool, log: log}

	app := fiber.New(fiber.Config{
		AppName:      "bdns-api",
		ErrorHandler: errorHandler,
	})
	a.routes(app)

	go func() {
		<-ctx.Done()
		_ = app.ShutdownWithTimeout(5 * time.Second)
	}()

	log.Info().Str("addr", addr).Msg("management API listening")
	if err := app.Listen(addr, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
		return err
	}
	return nil
}

// errorHandler renders any handler error (including fiber.Error) as JSON.
func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if errors.As(err, &fe) {
		code = fe.Code
	}
	return c.Status(code).JSON(fiber.Map{"error": err.Error()})
}

// auth is the bearer-token middleware applied to every protected route.
func (a *API) auth(c fiber.Ctx) error {
	token := bearerToken(c.Get(fiber.HeaderAuthorization))
	if token == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "missing bearer token")
	}
	user, err := a.repo.UserByToken(c.Context(), token)
	if err != nil {
		a.log.Error().Err(err).Msg("token lookup failed")
		return fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if user == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}
	c.Locals(userLocal, user)
	return c.Next()
}

func bearerToken(h string) string {
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// currentUser returns the authenticated user attached by the auth middleware.
func currentUser(c fiber.Ctx) *models.User {
	u, _ := c.Locals(userLocal).(*models.User)
	return u
}

// newToken returns a 128-bit hex token for a fresh user.
func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// bind decodes the body as JSON regardless of Content-Type.
func bind(c fiber.Ctx, dst any) error {
	if err := c.Bind().JSON(dst); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid JSON body: "+err.Error())
	}
	return nil
}

// ---- ownership guards ----

// ownedProfile returns the :pid param, 404/403-ing if it isn't the user's.
func (a *API) ownedProfile(c fiber.Ctx) (string, error) {
	pid := c.Params("pid")
	profile, err := a.repo.GetProfile(c.Context(), pid)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if profile == nil {
		return "", fiber.NewError(fiber.StatusNotFound, "profile not found")
	}
	if profile.UserID != currentUser(c).ID {
		return "", fiber.NewError(fiber.StatusForbidden, "not your profile")
	}
	return pid, nil
}

// managedPool returns the :poolID pool, requiring the user to be its creator.
func (a *API) managedPool(c fiber.Ctx) (*models.FriendPool, error) {
	pool, err := a.loadPool(c)
	if err != nil {
		return nil, err
	}
	if pool.CreatedBy != currentUser(c).ID {
		return nil, fiber.NewError(fiber.StatusForbidden, "only the pool creator can do that")
	}
	return pool, nil
}

// accessiblePool returns the :poolID pool, requiring the user to be creator or member.
func (a *API) accessiblePool(c fiber.Ctx) (*models.FriendPool, error) {
	pool, err := a.loadPool(c)
	if err != nil {
		return nil, err
	}
	uid := currentUser(c).ID
	if pool.CreatedBy == uid {
		return pool, nil
	}
	member, err := a.repo.UserInPool(c.Context(), pool.ID, uid)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if !member {
		return nil, fiber.NewError(fiber.StatusForbidden, "not a member of this pool")
	}
	return pool, nil
}

func (a *API) loadPool(c fiber.Ctx) (*models.FriendPool, error) {
	pool, err := a.repo.GetPool(c.Context(), c.Params("poolID"))
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "internal error")
	}
	if pool == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "pool not found")
	}
	return pool, nil
}
