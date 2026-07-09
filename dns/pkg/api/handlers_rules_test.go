package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/will-x86/bdns/dns/pkg/db/models"
)

func TestPermanentWhitelist(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	pid := ta.newProfile(u, "laptop")
	base := "/api/v1/profiles/" + pid + "/whitelist/permanent"

	if got := ta.status(http.MethodPost, base, u.Token, map[string]string{"domain": " "}); got != http.StatusBadRequest {
		t.Fatalf("blank domain: want 400, got %d", got)
	}

	resp, body := ta.req(http.MethodPost, base, u.Token, map[string]string{"domain": "good.example.com"})
	assertStatus(t, resp, body, http.StatusNoContent)

	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.PermanentWhitelist](t, body); len(l) != 1 || l[0].Domain != "good.example.com" {
		t.Fatalf("list = %s", body)
	}

	resp, body = ta.req(http.MethodDelete, base+"/good.example.com", u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.PermanentWhitelist](t, body); len(l) != 0 {
		t.Fatalf("not deleted: %s", body)
	}
}

func TestTemporaryWhitelist(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	pid := ta.newProfile(u, "laptop")
	base := "/api/v1/profiles/" + pid + "/whitelist/temporary"

	// ttl_seconds
	resp, body := ta.req(http.MethodPost, base, u.Token, map[string]any{"domain": "a.example.com", "ttl_seconds": 3600})
	assertStatus(t, resp, body, http.StatusNoContent)
	// absolute expires_at
	future := time.Now().Add(time.Hour).Unix()
	resp, body = ta.req(http.MethodPost, base, u.Token, map[string]any{"domain": "b.example.com", "expires_at": future})
	assertStatus(t, resp, body, http.StatusNoContent)

	// past expiry rejected
	if got := ta.status(http.MethodPost, base, u.Token, map[string]any{"domain": "c.example.com", "expires_at": 1}); got != http.StatusBadRequest {
		t.Fatalf("past expiry: want 400, got %d", got)
	}
	// no expiry info rejected
	if got := ta.status(http.MethodPost, base, u.Token, map[string]any{"domain": "d.example.com"}); got != http.StatusBadRequest {
		t.Fatalf("no expiry: want 400, got %d", got)
	}

	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.TemporaryWhitelist](t, body); len(l) != 2 {
		t.Fatalf("list len = %d: %s", len(l), body)
	}

	resp, body = ta.req(http.MethodDelete, base+"/a.example.com", u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.TemporaryWhitelist](t, body); len(l) != 1 {
		t.Fatalf("delete failed: %s", body)
	}
}

func TestCategoryBlocks(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	pid := ta.newProfile(u, "laptop")
	base := "/api/v1/profiles/" + pid + "/category-blocks"

	resp, body := ta.req(http.MethodPost, base, u.Token, map[string]string{"category": "social"})
	assertStatus(t, resp, body, http.StatusNoContent)

	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.CategoryBlock](t, body); len(l) != 1 || l[0].Category != "social" {
		t.Fatalf("list = %s", body)
	}

	resp, body = ta.req(http.MethodDelete, base+"/social", u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.CategoryBlock](t, body); len(l) != 0 {
		t.Fatalf("not deleted: %s", body)
	}
}

func TestTimeBlocks(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	pid := ta.newProfile(u, "laptop")
	base := "/api/v1/profiles/" + pid + "/time-blocks"

	// valid
	resp, body := ta.req(http.MethodPost, base, u.Token, map[string]any{
		"category": "social", "start_time": 40, "end_time": 95, "day": 1,
	})
	assertStatus(t, resp, body, http.StatusNoContent)

	// invalid: slot out of range, end<start, bad day
	bad := []map[string]any{
		{"category": "social", "start_time": -1, "end_time": 5, "day": 1},
		{"category": "social", "start_time": 10, "end_time": 96, "day": 1},
		{"category": "social", "start_time": 10, "end_time": 5, "day": 1},
		{"category": "social", "start_time": 10, "end_time": 20, "day": 8},
		{"category": " ", "start_time": 10, "end_time": 20, "day": 1},
	}
	for i, b := range bad {
		if got := ta.status(http.MethodPost, base, u.Token, b); got != http.StatusBadRequest {
			t.Fatalf("bad time block %d: want 400, got %d", i, got)
		}
	}

	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.TimeBlock](t, body); len(l) != 1 {
		t.Fatalf("list len = %d: %s", len(l), body)
	}

	// delete by query
	resp, body = ta.req(http.MethodDelete, base+"?category=social&day=1&start_time=40&end_time=95", u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.TimeBlock](t, body); len(l) != 0 {
		t.Fatalf("not deleted: %s", body)
	}

	// delete missing params
	if got := ta.status(http.MethodDelete, base+"?category=social&day=1", u.Token, nil); got != http.StatusBadRequest {
		t.Fatalf("delete missing params: want 400, got %d", got)
	}
}

func TestCategories(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()

	_, body := ta.req(http.MethodGet, "/api/v1/categories", u.Token, nil)
	if l := decode[[]string](t, body); len(l) != 0 {
		t.Fatalf("expected empty categories, got %s", body)
	}

	execSQL(t, `INSERT INTO blocklist_sources (name, category) VALUES ('src', 'ads')`)
	execSQL(t, `INSERT INTO blocklist_sources (name, category) VALUES ('src2', 'social')`)

	_, body = ta.req(http.MethodGet, "/api/v1/categories", u.Token, nil)
	l := decode[[]string](t, body)
	if len(l) != 2 || l[0] != "ads" || l[1] != "social" {
		t.Fatalf("categories = %s", body)
	}
}
