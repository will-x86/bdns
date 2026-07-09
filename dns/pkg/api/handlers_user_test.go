package api

import (
	"net/http"
	"testing"

	"github.com/will-x86/bdns/dns/pkg/db/models"
)

func TestHealth(t *testing.T) {
	ta := newTestAPI(t)
	resp, body := ta.req(http.MethodGet, "/api/v1/health", "", nil)
	assertStatus(t, resp, body, http.StatusOK)
}

func TestAuthRequired(t *testing.T) {
	ta := newTestAPI(t)

	if got := ta.status(http.MethodGet, "/api/v1/profiles", "", nil); got != http.StatusUnauthorized {
		t.Fatalf("no token: want 401, got %d", got)
	}
	if got := ta.status(http.MethodGet, "/api/v1/profiles", "deadbeef", nil); got != http.StatusUnauthorized {
		t.Fatalf("bad token: want 401, got %d", got)
	}
}

func TestCreateUserValidation(t *testing.T) {
	ta := newTestAPI(t)
	if got := ta.status(http.MethodPost, "/api/v1/users", "", map[string]string{"timezone": "  "}); got != http.StatusBadRequest {
		t.Fatalf("blank timezone: want 400, got %d", got)
	}
	resp, body := ta.reqRaw(http.MethodPost, "/api/v1/users", "", []byte("{not json"))
	assertStatus(t, resp, body, http.StatusBadRequest)
}

func TestUserLifecycle(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()

	// me
	resp, body := ta.req(http.MethodGet, "/api/v1/me", u.Token, nil)
	assertStatus(t, resp, body, http.StatusOK)
	me := decode[models.User](t, body)
	if me.ID != u.ID {
		t.Fatalf("me id = %q, want %q", me.ID, u.ID)
	}
	if me.APIToken != "" {
		t.Fatalf("me leaked api_token: %s", body)
	}

	// update timezone
	resp, body = ta.req(http.MethodPatch, "/api/v1/me", u.Token, map[string]string{"timezone": "America/New_York"})
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, "/api/v1/me", u.Token, nil)
	if tz := decode[models.User](t, body).Timezone; tz != "America/New_York" {
		t.Fatalf("timezone = %q, want updated", tz)
	}

	// rotate token: old invalid, new valid
	resp, body = ta.req(http.MethodPost, "/api/v1/me/token", u.Token, nil)
	assertStatus(t, resp, body, http.StatusOK)
	newTok := decode[struct {
		APIToken string `json:"api_token"`
	}](t, body).APIToken
	if newTok == "" || newTok == u.Token {
		t.Fatalf("token not rotated: %s", body)
	}
	if got := ta.status(http.MethodGet, "/api/v1/me", u.Token, nil); got != http.StatusUnauthorized {
		t.Fatalf("old token still works: %d", got)
	}
	if got := ta.status(http.MethodGet, "/api/v1/me", newTok, nil); got != http.StatusOK {
		t.Fatalf("new token rejected: %d", got)
	}

	// delete
	resp, body = ta.req(http.MethodDelete, "/api/v1/me", newTok, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	if got := ta.status(http.MethodGet, "/api/v1/me", newTok, nil); got != http.StatusUnauthorized {
		t.Fatalf("deleted user token still works: %d", got)
	}
}

func TestFriends(t *testing.T) {
	ta := newTestAPI(t)
	a := ta.newUser()
	b := ta.newUser()

	// self-friend rejected
	if got := ta.status(http.MethodPost, "/api/v1/friends", a.Token, map[string]string{"friend_id": a.ID}); got != http.StatusBadRequest {
		t.Fatalf("self-friend: want 400, got %d", got)
	}
	// unknown friend
	if got := ta.status(http.MethodPost, "/api/v1/friends", a.Token, map[string]string{"friend_id": "nope"}); got != http.StatusNotFound {
		t.Fatalf("unknown friend: want 404, got %d", got)
	}

	// add (bidirectional)
	resp, body := ta.req(http.MethodPost, "/api/v1/friends", a.Token, map[string]string{"friend_id": b.ID})
	assertStatus(t, resp, body, http.StatusNoContent)

	_, body = ta.req(http.MethodGet, "/api/v1/friends", a.Token, nil)
	if fs := decode[[]models.User](t, body); len(fs) != 1 || fs[0].ID != b.ID {
		t.Fatalf("A friends = %s", body)
	}
	_, body = ta.req(http.MethodGet, "/api/v1/friends", b.Token, nil)
	if fs := decode[[]models.User](t, body); len(fs) != 1 || fs[0].ID != a.ID {
		t.Fatalf("B friends (bidirectional) = %s", body)
	}

	// delete removes both directions
	resp, body = ta.req(http.MethodDelete, "/api/v1/friends/"+b.ID, a.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, "/api/v1/friends", b.Token, nil)
	if fs := decode[[]models.User](t, body); len(fs) != 0 {
		t.Fatalf("friendship not removed for B: %s", body)
	}
}

func TestProfileCRUD(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()

	// missing name
	if got := ta.status(http.MethodPost, "/api/v1/profiles", u.Token, map[string]string{"name": " "}); got != http.StatusBadRequest {
		t.Fatalf("blank name: want 400, got %d", got)
	}

	pid := ta.newProfile(u, "laptop")

	_, body := ta.req(http.MethodGet, "/api/v1/profiles", u.Token, nil)
	if ps := decode[[]models.Profile](t, body); len(ps) != 1 || ps[0].ID != pid {
		t.Fatalf("list = %s", body)
	}

	resp, body := ta.req(http.MethodGet, "/api/v1/profiles/"+pid, u.Token, nil)
	assertStatus(t, resp, body, http.StatusOK)

	resp, body = ta.req(http.MethodPatch, "/api/v1/profiles/"+pid, u.Token, map[string]string{"name": "phone"})
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, "/api/v1/profiles/"+pid, u.Token, nil)
	if name := decode[models.Profile](t, body).Name; name != "phone" {
		t.Fatalf("name = %q, want phone", name)
	}

	resp, body = ta.req(http.MethodDelete, "/api/v1/profiles/"+pid, u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	if got := ta.status(http.MethodGet, "/api/v1/profiles/"+pid, u.Token, nil); got != http.StatusNotFound {
		t.Fatalf("deleted profile: want 404, got %d", got)
	}
}

func TestProfileOwnership(t *testing.T) {
	ta := newTestAPI(t)
	a := ta.newUser()
	b := ta.newUser()
	pid := ta.newProfile(a, "a-laptop")

	// B cannot see or mutate A's profile
	if got := ta.status(http.MethodGet, "/api/v1/profiles/"+pid, b.Token, nil); got != http.StatusForbidden {
		t.Fatalf("B get A profile: want 403, got %d", got)
	}
	if got := ta.status(http.MethodPatch, "/api/v1/profiles/"+pid, b.Token, map[string]string{"name": "x"}); got != http.StatusForbidden {
		t.Fatalf("B patch A profile: want 403, got %d", got)
	}
	if got := ta.status(http.MethodDelete, "/api/v1/profiles/"+pid, b.Token, nil); got != http.StatusForbidden {
		t.Fatalf("B delete A profile: want 403, got %d", got)
	}
	// nonexistent profile
	if got := ta.status(http.MethodGet, "/api/v1/profiles/nope", a.Token, nil); got != http.StatusNotFound {
		t.Fatalf("missing profile: want 404, got %d", got)
	}
}
