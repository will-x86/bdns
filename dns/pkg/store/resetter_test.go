package store

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/db"
)

// fakePool is a Pool whose Reset methods actually work (unlike PoolMemory), so
// the resetter's reset/skip logic can be exercised. It records reset calls.
type fakePool struct {
	shared           map[string]int64
	borrow           map[string]int64
	resetSharedCalls int
	resetBorrowCalls int
}

func newFakePool() *fakePool {
	return &fakePool{shared: map[string]int64{}, borrow: map[string]int64{}}
}

func borrowKey(poolID, profileID string) string { return poolID + ":" + profileID }

func (f *fakePool) PoolID(context.Context, string) (string, error) { return "", nil }

func (f *fakePool) ExistsShared(_ context.Context, poolID string) bool {
	_, ok := f.shared[poolID]
	return ok
}

func (f *fakePool) ExistsBorrow(_ context.Context, poolID, profileID string) bool {
	_, ok := f.borrow[borrowKey(poolID, profileID)]
	return ok
}

func (f *fakePool) GetRemainingShared(_ context.Context, poolID string) (int64, error) {
	return f.shared[poolID], nil
}

func (f *fakePool) GetRemainingBorrow(_ context.Context, poolID, profileID string) (int64, error) {
	return f.borrow[borrowKey(poolID, profileID)], nil
}

func (f *fakePool) DecrementRemainingBorrow(context.Context, string, string) error { return nil }
func (f *fakePool) DecrementRemainingShared(context.Context, string) error         { return nil }

func (f *fakePool) ResetShared(_ context.Context, poolID string, limit, _ int64) error {
	f.shared[poolID] = limit
	f.resetSharedCalls++
	return nil
}

func (f *fakePool) ResetBorrow(_ context.Context, poolID, profileID string, limit, _ int64) error {
	f.borrow[borrowKey(poolID, profileID)] = limit
	f.resetBorrowCalls++
	return nil
}

func initResetterDB(t *testing.T) *db.SQLiteStores {
	t.Helper()
	if err := db.InitDB(zerolog.Nop(), ":memory:", "../../migrations/"); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return db.NewStores(db.GetDB())
}

func TestResetter_ResetsAndSkips(t *testing.T) {
	stores := initResetterDB(t)
	dbx := db.GetDB()

	creator, _ := db.CreateUser("UTC")
	other, _ := db.CreateUser("UTC")
	cProfile, _ := db.CreateProfile(creator, "creator-laptop")
	oProfile, _ := db.CreateProfile(other, "other-laptop")

	// shared pool where the creator is a member -> should reset
	dbx.MustExec(`INSERT INTO friend_pools (id, created_by, name, pool_mode, total_limit) VALUES ('sp', ?, 'S', 'shared', 100)`, creator)
	dbx.MustExec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('sp', ?)`, cProfile)

	// shared pool where the creator is NOT a member -> never reset (created_by != member's user)
	dbx.MustExec(`INSERT INTO friend_pools (id, created_by, name, pool_mode, total_limit) VALUES ('sp2', ?, 'S2', 'shared', 500)`, creator)
	dbx.MustExec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('sp2', ?)`, oProfile)

	// borrow pool -> resets per member
	dbx.MustExec(`INSERT INTO friend_pools (id, created_by, name, pool_mode, total_limit) VALUES ('bp', ?, 'B', 'borrow', 200)`, creator)
	dbx.MustExec(`INSERT INTO friend_pool_members (pool_id, profile_id) VALUES ('bp', ?)`, cProfile)

	fp := newFakePool()
	r := NewResetter(stores, fp)
	r.resetPools(context.Background())

	if fp.shared["sp"] != 100 {
		t.Errorf("shared sp = %d, want 100", fp.shared["sp"])
	}
	if _, ok := fp.shared["sp2"]; ok {
		t.Error("sp2 should not be reset (creator is not a member)")
	}
	if got := fp.borrow[borrowKey("bp", cProfile)]; got != 200 {
		t.Errorf("borrow bp = %d, want 200", got)
	}
	if fp.resetSharedCalls != 1 {
		t.Errorf("resetSharedCalls = %d, want 1", fp.resetSharedCalls)
	}
	if fp.resetBorrowCalls != 1 {
		t.Errorf("resetBorrowCalls = %d, want 1", fp.resetBorrowCalls)
	}

	// second run: keys now exist, so nothing should be reset again
	r.resetPools(context.Background())
	if fp.resetSharedCalls != 1 || fp.resetBorrowCalls != 1 {
		t.Errorf("second run reset again: shared=%d borrow=%d, want 1/1", fp.resetSharedCalls, fp.resetBorrowCalls)
	}
}
