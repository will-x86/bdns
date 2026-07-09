package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/store"
)

// testAPI is a running Fiber app backed by a fresh migrated SQLite DB and an
// in-memory pool store.
type testAPI struct {
	t    *testing.T
	app  *fiber.App
	pool *store.PoolMemory
}

func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	if err := db.InitDB(zerolog.Nop(), dbPath, "../../migrations"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	mem := store.NewMemory()
	a := &API{repo: db.NewRepo(db.GetDB()), pool: mem, log: zerolog.Nop()}
	app := fiber.New(fiber.Config{ErrorHandler: errorHandler})
	a.routes(app)
	return &testAPI{t: t, app: app, pool: mem}
}

// req sends a request with a JSON-marshalled body (nil for none).
func (ta *testAPI) req(method, path, token string, body any) (*http.Response, []byte) {
	ta.t.Helper()
	var raw []byte
	if body != nil {
		var err error
		if raw, err = json.Marshal(body); err != nil {
			ta.t.Fatalf("marshal body: %v", err)
		}
	}
	return ta.reqRaw(method, path, token, raw)
}

func (ta *testAPI) reqRaw(method, path, token string, body []byte) (*http.Response, []byte) {
	ta.t.Helper()
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	httpReq := httptest.NewRequest(method, path, r)
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := ta.app.Test(httpReq)
	if err != nil {
		ta.t.Fatalf("%s %s: %v", method, path, err)
	}
	data, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, data
}

func (ta *testAPI) status(method, path, token string, body any) int {
	resp, _ := ta.req(method, path, token, body)
	return resp.StatusCode
}

// assertStatus fails the test (dumping the body) if the status doesn't match.
func assertStatus(t *testing.T, resp *http.Response, body []byte, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("want status %d, got %d: %s", want, resp.StatusCode, body)
	}
}

func decode[T any](t *testing.T, body []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return v
}

// user is a created user plus its token.
type user struct {
	ID       string
	Timezone string
	Token    string
}

func (ta *testAPI) newUser() user {
	ta.t.Helper()
	resp, body := ta.req(http.MethodPost, "/api/v1/users", "", map[string]string{"timezone": "Europe/London"})
	assertStatus(ta.t, resp, body, http.StatusCreated)
	out := decode[struct {
		ID       string `json:"id"`
		Timezone string `json:"timezone"`
		APIToken string `json:"api_token"`
	}](ta.t, body)
	if out.ID == "" || out.APIToken == "" {
		ta.t.Fatalf("empty id/token in %s", body)
	}
	return user{ID: out.ID, Timezone: out.Timezone, Token: out.APIToken}
}

func (ta *testAPI) newProfile(u user, name string) string {
	ta.t.Helper()
	resp, body := ta.req(http.MethodPost, "/api/v1/profiles", u.Token, map[string]string{"name": name})
	assertStatus(ta.t, resp, body, http.StatusCreated)
	return decode[struct {
		ID string `json:"id"`
	}](ta.t, body).ID
}

func (ta *testAPI) newPool(u user, name, mode string, limit int64) string {
	ta.t.Helper()
	resp, body := ta.req(http.MethodPost, "/api/v1/pools", u.Token, map[string]any{
		"name": name, "pool_mode": mode, "total_limit": limit,
	})
	assertStatus(ta.t, resp, body, http.StatusCreated)
	return decode[struct {
		ID string `json:"id"`
	}](ta.t, body).ID
}

// execSQL runs a statement against the shared DB (for seeding blocklist rows).
func execSQL(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := db.GetDB().Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
