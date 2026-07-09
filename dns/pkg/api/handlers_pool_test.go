package api

import (
	"net/http"
	"testing"

	"github.com/will-x86/bdns/dns/pkg/db/models"
)

func TestPoolCRUD(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()

	// validation
	if got := ta.status(http.MethodPost, "/api/v1/pools", u.Token, map[string]any{"name": "x", "pool_mode": "bogus"}); got != http.StatusBadRequest {
		t.Fatalf("bad mode: want 400, got %d", got)
	}
	if got := ta.status(http.MethodPost, "/api/v1/pools", u.Token, map[string]any{"name": " ", "pool_mode": "shared"}); got != http.StatusBadRequest {
		t.Fatalf("blank name: want 400, got %d", got)
	}

	// default limit applied when omitted
	resp, body := ta.req(http.MethodPost, "/api/v1/pools", u.Token, map[string]any{"name": "Pot", "pool_mode": "shared"})
	assertStatus(t, resp, body, http.StatusCreated)
	created := decode[models.FriendPool](t, body)
	if created.TotalLimit != 6000 {
		t.Fatalf("default limit = %d, want 6000", created.TotalLimit)
	}

	pid := created.ID
	_, body = ta.req(http.MethodGet, "/api/v1/pools", u.Token, nil)
	if l := decode[[]models.FriendPool](t, body); len(l) != 1 || l[0].ID != pid {
		t.Fatalf("list = %s", body)
	}

	resp, body = ta.req(http.MethodPatch, "/api/v1/pools/"+pid, u.Token, map[string]any{"name": "Renamed", "total_limit": 1234})
	assertStatus(t, resp, body, http.StatusNoContent)
	_, body = ta.req(http.MethodGet, "/api/v1/pools/"+pid, u.Token, nil)
	got := decode[models.FriendPool](t, body)
	if got.Name != "Renamed" || got.TotalLimit != 1234 {
		t.Fatalf("update not applied: %s", body)
	}

	resp, body = ta.req(http.MethodDelete, "/api/v1/pools/"+pid, u.Token, nil)
	assertStatus(t, resp, body, http.StatusNoContent)
	if s := ta.status(http.MethodGet, "/api/v1/pools/"+pid, u.Token, nil); s != http.StatusNotFound {
		t.Fatalf("deleted pool: want 404, got %d", s)
	}
}

func TestPoolAccessControl(t *testing.T) {
	ta := newTestAPI(t)
	creator := ta.newUser()
	member := ta.newUser()
	stranger := ta.newUser()

	// befriend + add member's profile so member has access
	memberProfile := ta.newProfile(member, "m")
	if s := ta.status(http.MethodPost, "/api/v1/friends", creator.Token, map[string]string{"friend_id": member.ID}); s != http.StatusNoContent {
		t.Fatalf("befriend: %d", s)
	}
	poolID := ta.newPool(creator, "Social", "shared", 100)
	if s := ta.status(http.MethodPost, "/api/v1/pools/"+poolID+"/members", creator.Token, map[string]string{"profile_id": memberProfile}); s != http.StatusNoContent {
		t.Fatalf("add member: %d", s)
	}

	// member can read
	if s := ta.status(http.MethodGet, "/api/v1/pools/"+poolID, member.Token, nil); s != http.StatusOK {
		t.Fatalf("member read: want 200, got %d", s)
	}
	// member appears in their pool list
	_, body := ta.req(http.MethodGet, "/api/v1/pools", member.Token, nil)
	if l := decode[[]models.FriendPool](t, body); len(l) != 1 || l[0].ID != poolID {
		t.Fatalf("member pool list = %s", body)
	}
	// member cannot mutate
	if s := ta.status(http.MethodDelete, "/api/v1/pools/"+poolID, member.Token, nil); s != http.StatusForbidden {
		t.Fatalf("member delete: want 403, got %d", s)
	}
	if s := ta.status(http.MethodPatch, "/api/v1/pools/"+poolID, member.Token, map[string]any{"name": "x"}); s != http.StatusForbidden {
		t.Fatalf("member patch: want 403, got %d", s)
	}
	// stranger cannot read
	if s := ta.status(http.MethodGet, "/api/v1/pools/"+poolID, stranger.Token, nil); s != http.StatusForbidden {
		t.Fatalf("stranger read: want 403, got %d", s)
	}
}

func TestPoolMembers(t *testing.T) {
	ta := newTestAPI(t)
	creator := ta.newUser()
	friend := ta.newUser()
	stranger := ta.newUser()

	ownProfile := ta.newProfile(creator, "own")
	friendProfile := ta.newProfile(friend, "friends")
	strangerProfile := ta.newProfile(stranger, "stranger")

	if s := ta.status(http.MethodPost, "/api/v1/friends", creator.Token, map[string]string{"friend_id": friend.ID}); s != http.StatusNoContent {
		t.Fatalf("befriend: %d", s)
	}
	poolID := ta.newPool(creator, "Borrow", "borrow", 3000)
	base := "/api/v1/pools/" + poolID + "/members"

	// own + friend allowed
	if s := ta.status(http.MethodPost, base, creator.Token, map[string]string{"profile_id": ownProfile}); s != http.StatusNoContent {
		t.Fatalf("add own: %d", s)
	}
	if s := ta.status(http.MethodPost, base, creator.Token, map[string]string{"profile_id": friendProfile}); s != http.StatusNoContent {
		t.Fatalf("add friend: %d", s)
	}
	// stranger's profile forbidden
	if s := ta.status(http.MethodPost, base, creator.Token, map[string]string{"profile_id": strangerProfile}); s != http.StatusForbidden {
		t.Fatalf("add stranger: want 403, got %d", s)
	}
	// nonexistent profile
	if s := ta.status(http.MethodPost, base, creator.Token, map[string]string{"profile_id": "nope"}); s != http.StatusNotFound {
		t.Fatalf("add missing: want 404, got %d", s)
	}
	// non-creator cannot add
	if s := ta.status(http.MethodPost, base, friend.Token, map[string]string{"profile_id": friendProfile}); s != http.StatusForbidden {
		t.Fatalf("friend add member: want 403, got %d", s)
	}

	_, body := ta.req(http.MethodGet, base, creator.Token, nil)
	if l := decode[[]models.FriendPoolMembers](t, body); len(l) != 2 {
		t.Fatalf("members len = %d: %s", len(l), body)
	}

	if s := ta.status(http.MethodDelete, base+"/"+ownProfile, creator.Token, nil); s != http.StatusNoContent {
		t.Fatalf("remove member: %d", s)
	}
	_, body = ta.req(http.MethodGet, base, creator.Token, nil)
	if l := decode[[]models.FriendPoolMembers](t, body); len(l) != 1 {
		t.Fatalf("remove failed: %s", body)
	}
}

func TestPoolCategoryBlocks(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	poolID := ta.newPool(u, "Social", "shared", 100)
	base := "/api/v1/pools/" + poolID + "/category-blocks"

	if s := ta.status(http.MethodPost, base, u.Token, map[string]string{"category": "social"}); s != http.StatusNoContent {
		t.Fatalf("add: %d", s)
	}
	_, body := ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.FriendPoolCategoryBlocks](t, body); len(l) != 1 || l[0].Category != "social" {
		t.Fatalf("list = %s", body)
	}
	if s := ta.status(http.MethodDelete, base+"/social", u.Token, nil); s != http.StatusNoContent {
		t.Fatalf("delete: %d", s)
	}
	_, body = ta.req(http.MethodGet, base, u.Token, nil)
	if l := decode[[]models.FriendPoolCategoryBlocks](t, body); len(l) != 0 {
		t.Fatalf("not deleted: %s", body)
	}
}

func TestPoolLimitsShared(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	poolID := ta.newPool(u, "Pot", "shared", 100)

	// unknown until store is seeded
	_, body := ta.req(http.MethodGet, "/api/v1/pools/"+poolID+"/limits", u.Token, nil)
	got := decode[poolLimitsResp](t, body)
	if got.Mode != "shared" || got.Remaining != nil {
		t.Fatalf("pre-seed limits = %s", body)
	}

	ta.pool.SetShared(poolID, 42)
	_, body = ta.req(http.MethodGet, "/api/v1/pools/"+poolID+"/limits", u.Token, nil)
	got = decode[poolLimitsResp](t, body)
	if got.Remaining == nil || *got.Remaining != 42 || got.TotalLimit != 100 {
		t.Fatalf("post-seed limits = %s", body)
	}
}

func TestPoolLimitsBorrow(t *testing.T) {
	ta := newTestAPI(t)
	u := ta.newUser()
	profileA := ta.newProfile(u, "a")
	profileB := ta.newProfile(u, "b")
	poolID := ta.newPool(u, "Borrow", "borrow", 3000)

	for _, pr := range []string{profileA, profileB} {
		if s := ta.status(http.MethodPost, "/api/v1/pools/"+poolID+"/members", u.Token, map[string]string{"profile_id": pr}); s != http.StatusNoContent {
			t.Fatalf("add member %s: %d", pr, s)
		}
	}
	ta.pool.SetBorrow(poolID, profileA, 1500) // only A seeded

	_, body := ta.req(http.MethodGet, "/api/v1/pools/"+poolID+"/limits", u.Token, nil)
	got := decode[poolLimitsResp](t, body)
	if got.Mode != "borrow" || len(got.Members) != 2 {
		t.Fatalf("borrow limits = %s", body)
	}
	byProfile := map[string]*int64{}
	for _, m := range got.Members {
		byProfile[m.ProfileID] = m.Remaining
	}
	if byProfile[profileA] == nil || *byProfile[profileA] != 1500 {
		t.Fatalf("A remaining = %v, want 1500", byProfile[profileA])
	}
	if byProfile[profileB] != nil {
		t.Fatalf("B remaining = %v, want nil", *byProfile[profileB])
	}
}
