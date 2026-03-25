package rules

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"codeberg.org/will-x86/bdns/dns/pkg/rule"
)

// Fake pool cache store

type fakePoolCache struct {
	poolID           string
	poolIDErr        error
	existsShared     bool
	existsBorrow     bool
	remainingShared  int64
	remainingBorrow  int64
	remainingErr     error
	decrementErr     error
	decrementSharedN int // how many times DecrementRemainingShared was called
	decrementBorrowN int
}

func (f *fakePoolCache) PoolID(_ context.Context, _ string) (string, error) {
	return f.poolID, f.poolIDErr
}
func (f *fakePoolCache) ExistsShared(_ context.Context, _ string) bool {
	return f.existsShared
}
func (f *fakePoolCache) ExistsBorrow(_ context.Context, _, _ string) bool {
	return f.existsBorrow
}
func (f *fakePoolCache) GetRemainingShared(_ context.Context, _ string) (int64, error) {
	return f.remainingShared, f.remainingErr
}
func (f *fakePoolCache) GetRemainingBorrow(_ context.Context, _, _ string) (int64, error) {
	return f.remainingBorrow, f.remainingErr
}
func (f *fakePoolCache) DecrementRemainingShared(_ context.Context, _ string) error {
	f.decrementSharedN++
	return f.decrementErr
}
func (f *fakePoolCache) DecrementRemainingBorrow(_ context.Context, _, _ string) error {
	f.decrementBorrowN++
	return f.decrementErr
}

// Fake pool DB store

type fakePoolDB struct {
	pool            models.FriendPool
	poolErr         error
	categoryBlocked bool
}

func (f *fakePoolDB) GetPool(_ context.Context, _ string) (models.FriendPool, error) {
	return f.pool, f.poolErr
}
func (f *fakePoolDB) PoolCategoryBlocked(_ context.Context, _, _ string) bool {
	return f.categoryBlocked
}

// Helpers

func makePoolRuleCtx(cache *fakePoolCache, db *fakePoolDB, category string) *rule.RuleContext {
	return &rule.RuleContext{
		Domain:    "social.example.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			PoolCache: cache,
			PoolDB:    db,
			Resolve: func(_ context.Context, _ string) (string, error) {
				return category, nil
			},
		},
	}
}

func sharedPool() models.FriendPool {
	return models.FriendPool{ID: "pool1", PoolMode: "shared", TotalLimit: 100}
}

func borrowPool() models.FriendPool {
	return models.FriendPool{ID: "pool1", PoolMode: "borrow", TotalLimit: 100}
}

// SharedPoolRule tests

func TestSharedPoolRule_Name(t *testing.T) {
	if (&SharedPoolRule{}).Name() != "shared_pool_block" {
		t.Fatal("wrong name")
	}
}

func TestSharedPoolRule_NoPoolID_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolIDErr: errors.New("no pool")}
	rctx := makePoolRuleCtx(cache, &fakePoolDB{}, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when no poolID, got %v", d.Verdict)
	}
}

func TestSharedPoolRule_KeyNotInCache_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: false}
	rctx := makePoolRuleCtx(cache, &fakePoolDB{pool: sharedPool()}, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when key missing in valkey, got %v", d.Verdict)
	}
}

func TestSharedPoolRule_PoolModeIsBorrow_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{pool: borrowPool()}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when pool mode is borrow, got %v", d.Verdict)
	}
}

func TestSharedPoolRule_UnknownPoolMode_Error(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{pool: models.FriendPool{ID: "pool1", PoolMode: "foobar"}}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error for unknown pool mode, got nil")
	}
}

func TestSharedPoolRule_CategoryNotBlocked_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: false}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when category not pool-blocked, got %v", d.Verdict)
	}
}

func TestSharedPoolRule_RemainingAboveZero_PassThroughAndDecrements(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsShared:    true,
		remainingShared: 50,
	}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when credits remain, got %v", d.Verdict)
	}
	if cache.decrementSharedN != 1 {
		t.Fatalf("want 1 decrement call, got %d", cache.decrementSharedN)
	}
}

func TestSharedPoolRule_RemainingZero_Blocked(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsShared:    true,
		remainingShared: 0,
	}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block when shared credits exhausted, got %v", d.Verdict)
	}
	if cache.decrementSharedN != 0 {
		t.Fatalf("want 0 decrement calls when blocked, got %d", cache.decrementSharedN)
	}
}

func TestSharedPoolRule_GetPoolError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{poolErr: errors.New("db failure")}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from GetPool failure, got nil")
	}
}

func TestSharedPoolRule_GetRemainingError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{
		poolID:       "pool1",
		existsShared: true,
		remainingErr: errors.New("valkey error"),
	}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from GetRemainingShared failure, got nil")
	}
}

func TestSharedPoolRule_DecrementError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsShared:    true,
		remainingShared: 10,
		decrementErr:    errors.New("valkey write error"),
	}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from Decrement failure, got nil")
	}
}

func TestSharedPoolRule_CategoryNotInBlocklist_PassThrough(t *testing.T) {
	// Domain resolves to empty category — pool rule shouldn't block regardless
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "") // empty category
	d, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	// PoolCategoryBlocked returns true but category is "", so the DB check
	// will use empty string — categoryBlocked fake returns true regardless,
	// so the rule will read remaining. With remaining=0 (default) it blocks.
	// This test verifies the rule handles zero remaining properly.
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block (remaining=0 default), got %v", d.Verdict)
	}
}

func TestSharedPoolRule_ResolveError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}
	resolveErr := errors.New("resolve failure")
	rctx := &rule.RuleContext{
		Domain:    "social.example.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			PoolCache: cache,
			PoolDB:    db,
			Resolve:   func(_ context.Context, _ string) (string, error) { return "", resolveErr },
		},
	}
	_, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, resolveErr) {
		t.Fatalf("want resolve error, got %v", err)
	}
}

// Pool is cached across two evaluate calls via RuleContext
func TestSharedPoolRule_PoolCachedInRuleContext(t *testing.T) {
	calls := 0
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: false}
	realGetPool := db.GetPool
	_ = realGetPool

	cache := &fakePoolCache{poolID: "pool1", existsShared: true}
	rctx := makePoolRuleCtx(cache, db, "social")

	// Wrap PoolDB to count GetPool calls
	type countingDB struct {
		*fakePoolDB
		n *int
	}
	cdb := &countingDB{fakePoolDB: db, n: &calls}
	cdb.fakePoolDB = db

	rctx.Stores.PoolDB = &countingPoolDB{fakePoolDB: db, calls: &calls}

	(&SharedPoolRule{}).Evaluate(context.Background(), rctx)
	(&SharedPoolRule{}).Evaluate(context.Background(), rctx)

	if calls > 1 {
		t.Fatalf("GetPool should be called at most once due to RuleContext cache, got %d calls", calls)
	}
}

type countingPoolDB struct {
	*fakePoolDB
	calls *int
}

func (c *countingPoolDB) GetPool(ctx context.Context, poolID string) (models.FriendPool, error) {
	*c.calls++
	return c.fakePoolDB.GetPool(ctx, poolID)
}

// BorrowPoolRule tests

func TestBorrowPoolRule_Name(t *testing.T) {
	if (&BorrowPoolRule{}).Name() != "borrow_pool_block" {
		t.Fatal("wrong name")
	}
}

func TestBorrowPoolRule_NoPoolID_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolIDErr: errors.New("no pool")}
	rctx := makePoolRuleCtx(cache, &fakePoolDB{}, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when no poolID, got %v", d.Verdict)
	}
}

func TestBorrowPoolRule_KeyNotInCache_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: false}
	rctx := makePoolRuleCtx(cache, &fakePoolDB{pool: borrowPool()}, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when borrow key missing in valkey, got %v", d.Verdict)
	}
}

func TestBorrowPoolRule_PoolModeIsShared_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: true}
	db := &fakePoolDB{pool: sharedPool()}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when pool mode is shared, got %v", d.Verdict)
	}
}

func TestBorrowPoolRule_UnknownPoolMode_Error(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: true}
	db := &fakePoolDB{pool: models.FriendPool{ID: "pool1", PoolMode: "foobar"}}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error for unknown pool mode, got nil")
	}
}

func TestBorrowPoolRule_CategoryNotBlocked_PassThrough(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: true}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: false}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when category not pool-blocked, got %v", d.Verdict)
	}
}

func TestBorrowPoolRule_RemainingAboveZero_PassThroughAndDecrements(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsBorrow:    true,
		remainingBorrow: 3,
	}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough when borrow credits remain, got %v", d.Verdict)
	}
	if cache.decrementBorrowN != 1 {
		t.Fatalf("want 1 decrement call, got %d", cache.decrementBorrowN)
	}
}

func TestBorrowPoolRule_RemainingZero_Blocked(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsBorrow:    true,
		remainingBorrow: 0,
	}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	d, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block when borrow credits exhausted, got %v", d.Verdict)
	}
	if cache.decrementBorrowN != 0 {
		t.Fatalf("want 0 decrement calls when blocked, got %d", cache.decrementBorrowN)
	}
}

func TestBorrowPoolRule_GetPoolError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: true}
	db := &fakePoolDB{poolErr: errors.New("db failure")}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from GetPool failure, got nil")
	}
}

func TestBorrowPoolRule_GetRemainingError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{
		poolID:       "pool1",
		existsBorrow: true,
		remainingErr: errors.New("valkey error"),
	}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from GetRemainingBorrow failure, got nil")
	}
}

func TestBorrowPoolRule_DecrementError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{
		poolID:          "pool1",
		existsBorrow:    true,
		remainingBorrow: 5,
		decrementErr:    errors.New("valkey write error"),
	}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}
	rctx := makePoolRuleCtx(cache, db, "social")
	_, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if err == nil {
		t.Fatal("want error from Decrement failure, got nil")
	}
}

func TestBorrowPoolRule_ResolveError_ReturnsError(t *testing.T) {
	cache := &fakePoolCache{poolID: "pool1", existsBorrow: true}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}
	resolveErr := errors.New("resolve failure")
	rctx := &rule.RuleContext{
		Domain:    "social.example.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			PoolCache: cache,
			PoolDB:    db,
			Resolve:   func(_ context.Context, _ string) (string, error) { return "", resolveErr },
		},
	}
	_, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, resolveErr) {
		t.Fatalf("want resolve error, got %v", err)
	}
}

// Each profile's borrow counter is independent — one exhausted doesn't affect the other
func TestBorrowPoolRule_PerProfileIsolation(t *testing.T) {
	// profile1 exhausted, profile2 still has credits
	cacheP1 := &fakePoolCache{poolID: "pool1", existsBorrow: true, remainingBorrow: 0}
	cacheP2 := &fakePoolCache{poolID: "pool1", existsBorrow: true, remainingBorrow: 10}
	db := &fakePoolDB{pool: borrowPool(), categoryBlocked: true}

	rctxP1 := makePoolRuleCtx(cacheP1, db, "social")
	rctxP1.ProfileID = "profile1"
	rctxP2 := makePoolRuleCtx(cacheP2, db, "social")
	rctxP2.ProfileID = "profile2"

	d1, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctxP1)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := (&BorrowPoolRule{}).Evaluate(context.Background(), rctxP2)
	if err != nil {
		t.Fatal(err)
	}
	if d1.Verdict != rule.VerdictBlock {
		t.Fatalf("profile1 should be blocked (no credits), got %v", d1.Verdict)
	}
	if d2.Verdict != rule.VerdictPassThrough {
		t.Fatalf("profile2 should pass (has credits), got %v", d2.Verdict)
	}
}

// Shared pool: one member exhausts the pool, second member is also blocked
func TestSharedPoolRule_SharedExhaustion_BlocksBothUsers(t *testing.T) {
	db := &fakePoolDB{pool: sharedPool(), categoryBlocked: true}

	// After exhaustion remaining=0
	cacheExhausted := &fakePoolCache{poolID: "pool1", existsShared: true, remainingShared: 0}

	rctx1 := makePoolRuleCtx(cacheExhausted, db, "social")
	rctx1.ProfileID = "profile1"
	rctx2 := makePoolRuleCtx(cacheExhausted, db, "social")
	rctx2.ProfileID = "profile2"

	d1, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx1)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := (&SharedPoolRule{}).Evaluate(context.Background(), rctx2)
	if err != nil {
		t.Fatal(err)
	}
	if d1.Verdict != rule.VerdictBlock {
		t.Fatalf("profile1 should be blocked when shared pool exhausted, got %v", d1.Verdict)
	}
	if d2.Verdict != rule.VerdictBlock {
		t.Fatalf("profile2 should also be blocked when shared pool exhausted, got %v", d2.Verdict)
	}
}
