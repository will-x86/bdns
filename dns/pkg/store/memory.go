package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type PoolMemory struct {
	mu            sync.RWMutex
	profilepoolid map[string]string // profile_id -> pool_id
	shared        map[string]int64  // pool_id -> remaining shared credits
	borrow        map[string]int64  // "pool_id:profile_id" -> remaining borrow credits
}

func NewMemory() *PoolMemory {
	return &PoolMemory{
		profilepoolid: make(map[string]string),
		shared:        make(map[string]int64),
		borrow:        make(map[string]int64),
	}
}

func (p *PoolMemory) sharedKey(poolID string) string {
	return "pool:" + poolID + ":credits"
}

func (p *PoolMemory) borrowKey(poolID, profileID string) string {
	return fmt.Sprintf("pool:%s:%s:credits", poolID, profileID)
}

func (p *PoolMemory) SetShared(poolID string, credits int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.shared[p.sharedKey(poolID)] = credits
}

func (p *PoolMemory) SetBorrow(poolID, profileID string, credits int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.borrow[p.borrowKey(poolID, profileID)] = credits
}

func (p *PoolMemory) SetProfilePool(profileID, poolID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.profilepoolid[profileID] = poolID
}

func (p *PoolMemory) PoolID(_ context.Context, profileID string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.profilepoolid[profileID]
	if !ok {
		return "", errors.New("pool does not exist with given profile_id")
	}
	return v, nil
}

func (p *PoolMemory) Exists(_ context.Context, profileID, poolID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, sharedOk := p.shared[p.sharedKey(poolID)]
	_, borrowOk := p.borrow[p.borrowKey(poolID, profileID)]
	return sharedOk || borrowOk
}

func (p *PoolMemory) GetRemainingShared(_ context.Context, poolID string) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.shared[p.sharedKey(poolID)]
	if !ok {
		return 0, fmt.Errorf("shared pool not found: %s", poolID)
	}
	return v, nil
}

func (p *PoolMemory) GetRemainingBorrow(_ context.Context, poolID, profileID string) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.borrow[p.borrowKey(poolID, profileID)]
	if !ok {
		return 0, fmt.Errorf("borrow pool not found: %s / %s", poolID, profileID)
	}
	return v, nil
}

func (p *PoolMemory) DecrementRemainingBorrow(_ context.Context, poolID, profileID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := p.borrowKey(poolID, profileID)
	if _, ok := p.borrow[key]; !ok {
		return fmt.Errorf("borrow pool not found: %s / %s", poolID, profileID)
	}
	p.borrow[key]--
	return nil
}

func (p *PoolMemory) DecrementRemainingShared(_ context.Context, poolID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	key := p.sharedKey(poolID)
	if _, ok := p.shared[key]; !ok {
		return fmt.Errorf("shared pool not found: %s", poolID)
	}
	p.shared[key]--
	return nil
}

func (p *PoolMemory) ResetBorrow(ctx context.Context, poolID, profileID string, limit int64, ttlSeconds int64) error {
	return errors.New("reset borrow memory unimplemented")
}

func (p *PoolMemory) ResetShared(ctx context.Context, poolID string, limit int64, ttlSeconds int64) error {

	return errors.New("reset shared memory unimplemented")
}
