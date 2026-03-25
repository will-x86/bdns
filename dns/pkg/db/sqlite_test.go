package db

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

const migrationsDir = "../../migrations/"

func initTestDB(t *testing.T) *SQLiteStores {
	t.Helper()
	if err := InitDB(zerolog.Nop(), ":memory:", migrationsDir); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return NewStores(GetDB())
}

func TestProfileExists_NotFound(t *testing.T) {
	s := initTestDB(t)

	exists, err := s.ProfileExists(context.Background(), "doesnotexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for unknown profile, got true")
	}
}

func TestProfileExists_Found(t *testing.T) {
	s := initTestDB(t)

	userID, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	profileID, err := CreateProfile(userID, "test-laptop")
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	exists, err := s.ProfileExists(context.Background(), profileID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true for profile %q, got false", profileID)
	}
}

func TestCreateUser_ReturnsID(t *testing.T) {
	_ = initTestDB(t)

	id, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty ID from CreateUser")
	}
}

func TestCreateUser_UniqueIDs(t *testing.T) {
	_ = initTestDB(t)

	a, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser (1): %v", err)
	}
	b, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser (2): %v", err)
	}
	if a == b {
		t.Errorf("expected unique IDs, both returned %q", a)
	}
}

// GetProfileWithUser

func TestGetProfileWithUser_NotFound(t *testing.T) {
	s := initTestDB(t)
	profile, user, err := s.GetProfileWithUser(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile != nil || user != nil {
		t.Fatal("want nil profile and user for unknown profileID")
	}
}

func TestGetProfileWithUser_Found(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("Europe/London")
	profileID, _ := CreateProfile(userID, "laptop")

	profile, user, err := s.GetProfileWithUser(context.Background(), profileID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile == nil || user == nil {
		t.Fatal("want non-nil profile and user")
	}
	if profile.ID != profileID {
		t.Errorf("profile.ID: want %q, got %q", profileID, profile.ID)
	}
	if user.ID != userID {
		t.Errorf("user.ID: want %q, got %q", userID, user.ID)
	}
	if user.Timezone != "Europe/London" {
		t.Errorf("user.Timezone: want Europe/London, got %q", user.Timezone)
	}
}

// IsPermanentlyWhitelisted

func TestIsPermanentlyWhitelisted_NotFound(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")

	ok, err := s.IsPermanentlyWhitelisted(context.Background(), profileID, "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false for non-whitelisted domain")
	}
}

func TestIsPermanentlyWhitelisted_Found(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	GetDB().Exec(`INSERT INTO permanent_whitelists (profile_id, domain) VALUES (?, ?)`, profileID, "safe.com")

	ok, err := s.IsPermanentlyWhitelisted(context.Background(), profileID, "safe.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want true for whitelisted domain")
	}
}

func TestIsPermanentlyWhitelisted_WrongProfile(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	GetDB().Exec(`INSERT INTO permanent_whitelists (profile_id, domain) VALUES (?, ?)`, profileID, "safe.com")

	ok, err := s.IsPermanentlyWhitelisted(context.Background(), "other-profile", "safe.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false when querying a different profile")
	}
}

// IsTemporarilyWhitelisted

func TestIsTemporarilyWhitelisted_NotFound(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")

	ok, err := s.IsTemporarilyWhitelisted(context.Background(), profileID, "temp.com", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false for non-whitelisted domain")
	}
}

func TestIsTemporarilyWhitelisted_Valid(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	future := time.Now().Unix() + 9999
	GetDB().Exec(`INSERT INTO temporary_whitelists (profile_id, domain, expires_at) VALUES (?, ?, ?)`, profileID, "temp.com", future)

	ok, err := s.IsTemporarilyWhitelisted(context.Background(), profileID, "temp.com", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want true for valid (unexpired) temp whitelist")
	}
}

func TestIsTemporarilyWhitelisted_Expired(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	past := time.Now().Unix() - 1
	GetDB().Exec(`INSERT INTO temporary_whitelists (profile_id, domain, expires_at) VALUES (?, ?, ?)`, profileID, "temp.com", past)

	ok, err := s.IsTemporarilyWhitelisted(context.Background(), profileID, "temp.com", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false for expired temp whitelist")
	}
}

// IsCategoryBlocked

func TestIsCategoryBlocked_NotBlocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")

	ok, err := s.IsCategoryBlocked(context.Background(), profileID, "social")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false when category not blocked")
	}
}

func TestIsCategoryBlocked_Blocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	GetDB().Exec(`INSERT INTO user_category_blocks (profile_id, category) VALUES (?, ?)`, profileID, "porn")

	ok, err := s.IsCategoryBlocked(context.Background(), profileID, "porn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("want true when category is blocked")
	}
}

func TestIsCategoryBlocked_OtherCategoryNotBlocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	profileID, _ := CreateProfile(userID, "test")
	GetDB().Exec(`INSERT INTO user_category_blocks (profile_id, category) VALUES (?, ?)`, profileID, "porn")

	ok, err := s.IsCategoryBlocked(context.Background(), profileID, "social")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("want false for category that is not in the block list")
	}
}

// ResolveCategory

func TestResolveCategory_NotFound(t *testing.T) {
	s := initTestDB(t)
	cat, err := s.ResolveCategory(context.Background(), "unknown.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != "" {
		t.Errorf("want empty string for unknown domain, got %q", cat)
	}
}

func TestResolveCategory_Found(t *testing.T) {
	s := initTestDB(t)
	GetDB().Exec(`INSERT INTO blocklist_sources (id, name, url, category) VALUES ('src1', 'test', 'http://x.com', 'social')`)
	GetDB().Exec(`INSERT INTO blocklist_entries (domain, source_id, category) VALUES ('twitter.com', 'src1', 'social')`)

	cat, err := s.ResolveCategory(context.Background(), "twitter.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cat != "social" {
		t.Errorf("want 'social', got %q", cat)
	}
}

// GetPool / PoolCategoryBlocked

func seedPool(t *testing.T, poolID, createdBy, mode string, limit int64) {
	t.Helper()
	_, err := GetDB().Exec(
		`INSERT INTO friend_pools (id, created_by, name, pool_mode, total_limit) VALUES (?, ?, ?, ?, ?)`,
		poolID, createdBy, "Test Pool", mode, limit,
	)
	if err != nil {
		t.Fatalf("seedPool: %v", err)
	}
}

func TestGetPool_Found(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	seedPool(t, "pool1", userID, "shared", 6000)

	pool, err := s.GetPool(context.Background(), "pool1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pool.ID != "pool1" {
		t.Errorf("want pool1, got %q", pool.ID)
	}
	if pool.PoolMode != "shared" {
		t.Errorf("want shared, got %q", pool.PoolMode)
	}
	if pool.TotalLimit != 6000 {
		t.Errorf("want 6000, got %d", pool.TotalLimit)
	}
}

func TestGetPool_NotFound(t *testing.T) {
	s := initTestDB(t)
	_, err := s.GetPool(context.Background(), "ghost-pool")
	if err == nil {
		t.Fatal("want error for non-existent pool")
	}
}

func TestPoolCategoryBlocked_NotBlocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	seedPool(t, "pool1", userID, "shared", 100)

	blocked := s.PoolCategoryBlocked(context.Background(), "pool1", "social")
	if blocked {
		t.Error("want false when no category block configured for pool")
	}
}

func TestPoolCategoryBlocked_Blocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	seedPool(t, "pool1", userID, "shared", 100)
	GetDB().Exec(`INSERT INTO friend_pool_category_blocks (pool_id, category) VALUES ('pool1', 'social')`)

	blocked := s.PoolCategoryBlocked(context.Background(), "pool1", "social")
	if !blocked {
		t.Error("want true when category is blocked for pool")
	}
}

func TestPoolCategoryBlocked_OtherCategoryNotBlocked(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("UTC")
	seedPool(t, "pool1", userID, "shared", 100)
	GetDB().Exec(`INSERT INTO friend_pool_category_blocks (pool_id, category) VALUES ('pool1', 'social')`)

	blocked := s.PoolCategoryBlocked(context.Background(), "pool1", "gambling")
	if blocked {
		t.Error("want false for a different category")
	}
}

// GetAllPoolMembersWithTimezones

func TestGetAllPoolMembersWithTimezones_Empty(t *testing.T) {
	s := initTestDB(t)
	members, err := s.GetAllPoolMembersWithTimezones(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 0 {
		t.Errorf("want 0 members on empty DB, got %d", len(members))
	}
}

func TestGetAllPoolMembersWithTimezones_SharedPool(t *testing.T) {
	s := initTestDB(t)
	userID, _ := CreateUser("Europe/London")
	profileID, _ := CreateProfile(userID, "laptop")
	seedPool(t, "pool1", userID, "shared", 100)
	GetDB().Exec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('pool1', ?)`, profileID)

	members, err := s.GetAllPoolMembersWithTimezones(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("want 1 member, got %d", len(members))
	}
	m := members[0]
	if m.PoolID != "pool1" {
		t.Errorf("PoolID: want pool1, got %q", m.PoolID)
	}
	if m.ProfileID != profileID {
		t.Errorf("ProfileID: want %q, got %q", profileID, m.ProfileID)
	}
	if m.PoolMode != "shared" {
		t.Errorf("PoolMode: want shared, got %q", m.PoolMode)
	}
	if m.TotalLimit != 100 {
		t.Errorf("TotalLimit: want 100, got %d", m.TotalLimit)
	}
	if m.Timezone != "Europe/London" {
		t.Errorf("Timezone: want Europe/London, got %q", m.Timezone)
	}
	if m.CreatedBy != userID {
		t.Errorf("CreatedBy: want %q, got %q", userID, m.CreatedBy)
	}
}

func TestGetAllPoolMembersWithTimezones_MultipleMembers(t *testing.T) {
	s := initTestDB(t)
	u1, _ := CreateUser("Europe/London")
	u2, _ := CreateUser("America/New_York")
	p1, _ := CreateProfile(u1, "laptop1")
	p2, _ := CreateProfile(u2, "laptop2")
	seedPool(t, "pool1", u1, "borrow", 3000)
	GetDB().Exec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('pool1', ?)`, p1)
	GetDB().Exec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('pool1', ?)`, p2)

	members, err := s.GetAllPoolMembersWithTimezones(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("want 2 members, got %d", len(members))
	}
	// collect timezones to confirm both are present
	tzs := map[string]bool{}
	for _, m := range members {
		tzs[m.Timezone] = true
	}
	if !tzs["Europe/London"] {
		t.Error("expected Europe/London timezone in results")
	}
	if !tzs["America/New_York"] {
		t.Error("expected America/New_York timezone in results")
	}
}
