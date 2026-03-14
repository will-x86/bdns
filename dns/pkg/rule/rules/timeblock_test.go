package rules

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/will-x86/bdns/dns/pkg/db"
	"github.com/will-x86/bdns/dns/pkg/db/models"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

const migrationsDir = "../../../migrations/"

// helpers

func makeCtx(category string, store rule.TimeBlockStore, now time.Time, tz string) *rule.RuleContext {
	return &rule.RuleContext{
		Domain:    "example.com",
		ProfileID: "p1",
		Now:       now,
		User:      &models.User{Timezone: tz},
		Stores: rule.Stores{
			Resolve: func(_ context.Context, _ string) (string, error) {
				return category, nil
			},
			TimeBlock: store,
		},
	}
}

type fakeTimeBlockStore struct {
	blocks []models.TimeBlock
	err    error
}

func (f fakeTimeBlockStore) GetTimeBlocks(_ context.Context, _, _ string) ([]models.TimeBlock, error) {
	return f.blocks, f.err
}

// unit tests

func TestTimeBlockRule_Name(t *testing.T) {
	if (&TimeBlockRule{}).Name() != "timeblock_rule" {
		t.Fatal("wrong name")
	}
}

func TestTimeBlockRule_NoCategory_PassThrough(t *testing.T) {
	rctx := makeCtx("", fakeTimeBlockStore{}, time.Now(), "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough, got %v", d.Verdict)
	}
}

func TestTimeBlockRule_NoRows_PassThrough(t *testing.T) {
	rctx := makeCtx("social", fakeTimeBlockStore{err: sql.ErrNoRows}, time.Now(), "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough, got %v", d.Verdict)
	}
}

func TestTimeBlockRule_StoreError(t *testing.T) {
	storeErr := errors.New("db failure")
	rctx := makeCtx("social", fakeTimeBlockStore{err: storeErr}, time.Now(), "UTC")
	_, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, storeErr) {
		t.Fatalf("want store error, got %v", err)
	}
}

func TestTimeBlockRule_ResolveError(t *testing.T) {
	resolveErr := errors.New("resolve failure")
	rctx := &rule.RuleContext{
		Domain: "example.com", Now: time.Now(), User: &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			Resolve:   func(_ context.Context, _ string) (string, error) { return "", resolveErr },
			TimeBlock: fakeTimeBlockStore{},
		},
	}
	_, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, resolveErr) {
		t.Fatalf("want resolve error, got %v", err)
	}
}

func TestTimeBlockRule_EmptyBlocks_Allowed(t *testing.T) {
	rctx := makeCtx("social", fakeTimeBlockStore{}, time.Now(), "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow, got %v", d.Verdict)
	}
}

// now = 09:00 UTC -> interval 36; block covers 32–39 (08:00–10:00)
func TestTimeBlockRule_InsideBlock_Blocked(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block at 09:00 inside [08:00,10:00], got %v", d.Verdict)
	}
	if !strings.Contains(d.Reason, "social") {
		t.Fatalf("reason missing category: %q", d.Reason)
	}
}

// now = 07:59 UTC -> interval 31; block covers 32–39 (08:00–10:00)
func TestTimeBlockRule_BeforeBlock_Allowed(t *testing.T) {
	now := time.Date(2026, 1, 1, 7, 59, 0, 0, time.UTC)
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow at 07:59 before [08:00,10:00], got %v", d.Verdict)
	}
}

// now = 10:01 UTC -> interval 40; block covers 32–39 (08:00–10:00)
func TestTimeBlockRule_AfterBlock_Allowed(t *testing.T) {
	now := time.Date(2026, 1, 1, 10, 1, 0, 0, time.UTC)
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow at 10:01 after [08:00,10:00], got %v", d.Verdict)
	}
}

// Block boundary: now exactly at StartTime interval.
func TestTimeBlockRule_AtStartBoundary_Blocked(t *testing.T) {
	now := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC) // interval 32
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block at start boundary (08:00), got %v", d.Verdict)
	}
}

// Block boundary: now exactly at EndTime interval.
func TestTimeBlockRule_AtEndBoundary_Blocked(t *testing.T) {
	now := time.Date(2026, 1, 1, 9, 45, 0, 0, time.UTC) // interval 39
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "UTC")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block at end boundary (09:45), got %v", d.Verdict)
	}
}

// Timezone: block is 08:00–10:00 UTC stored, user is in America/New_York (UTC-5).
// now = 13:00 UTC = 08:00 NY -> interval 32 in NY -> inside block.
func TestTimeBlockRule_Timezone_InsideBlock(t *testing.T) {
	now := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "America/New_York")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block: 13:00 UTC = 08:00 NY inside [08:00,10:00], got %v", d.Verdict)
	}
}

// Same block but now = 12:59 UTC = 07:59 NY -> interval 31 -> outside block.
func TestTimeBlockRule_Timezone_OutsideBlock(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 59, 0, 0, time.UTC)
	blocks := []models.TimeBlock{{StartTime: 32, EndTime: 39, Category: "social"}}
	rctx := makeCtx("social", fakeTimeBlockStore{blocks: blocks}, now, "America/New_York")
	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow: 12:59 UTC = 07:59 NY outside [08:00,10:00], got %v", d.Verdict)
	}
}

// db int tests

func initTestDB(t *testing.T) *db.SQLiteStores {
	t.Helper()
	if err := db.InitDB(":memory:", migrationsDir); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db.NewStores(db.GetDB())
}

func seedTimeBlock(t *testing.T, tz string, category string, start, end int) (profileID string) {
	t.Helper()
	userID, err := db.CreateUser(tz)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	profileID, err = db.CreateProfile(userID, "test")
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}
	_, err = db.GetDB().Exec(
		`INSERT INTO user_time_blocks (profile_id, category, start_time, end_time, day) VALUES (?, ?, ?, ?, 0)`,
		profileID, category, start, end,
	)
	if err != nil {
		t.Fatalf("insert time block: %v", err)
	}
	return profileID
}

func dbRuleCtx(stores *db.SQLiteStores, profileID, category, tz string, now time.Time) *rule.RuleContext {
	return &rule.RuleContext{
		Domain:    "social.example.com",
		ProfileID: profileID,
		Now:       now,
		User:      &models.User{Timezone: tz},
		Stores: rule.Stores{
			Resolve:   func(_ context.Context, _ string) (string, error) { return category, nil },
			TimeBlock: stores,
		},
	}
}

// now = 09:00 UTC -> interval 36; DB row has start=32 end=39
func TestDB_TimeBlock_InsideBlock_Blocked(t *testing.T) {
	stores := initTestDB(t)
	profileID := seedTimeBlock(t, "UTC", "social", 32, 39)

	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	rctx := dbRuleCtx(stores, profileID, "social", "UTC", now)

	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block, got %v (%s)", d.Verdict, d.Reason)
	}
}

// now = 07:59 UTC -> interval 31; DB row has start=32 end=39
func TestDB_TimeBlock_BeforeBlock_Allowed(t *testing.T) {
	stores := initTestDB(t)
	profileID := seedTimeBlock(t, "UTC", "social", 32, 39)

	now := time.Date(2026, 1, 1, 7, 59, 0, 0, time.UTC)
	rctx := dbRuleCtx(stores, profileID, "social", "UTC", now)

	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow, got %v (%s)", d.Verdict, d.Reason)
	}
}

// now = 13:00 UTC = 08:00 America/New_York -> interval 32 in NY -> inside block
func TestDB_TimeBlock_Timezone_Blocked(t *testing.T) {
	stores := initTestDB(t)
	profileID := seedTimeBlock(t, "America/New_York", "social", 32, 39)

	now := time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	rctx := dbRuleCtx(stores, profileID, "social", "America/New_York", now)

	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block: 13:00 UTC = 08:00 NY inside [08:00,10:00], got %v (%s)", d.Verdict, d.Reason)
	}
}

// No DB rows for this profile -> Allowed
func TestDB_TimeBlock_NoRows_Allowed(t *testing.T) {
	stores := initTestDB(t)
	userID, _ := db.CreateUser("UTC")
	profileID, _ := db.CreateProfile(userID, "empty")

	now := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	rctx := dbRuleCtx(stores, profileID, "social", "UTC", now)

	d, err := (&TimeBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow for profile with no blocks, got %v", d.Verdict)
	}
}
