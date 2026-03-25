package store

import "context"

type Pool interface {
	Exists(ctx context.Context, profileID, poolID string) bool
	PoolID(ctx context.Context, profileID string) (string, error)
	GetRemainingShared(ctx context.Context, poolID string) (int64, error)
	DecrementRemainingBorrow(ctx context.Context, poolID, profileID string) error
	DecrementRemainingShared(ctx context.Context, poolID string) error
	GetRemainingBorrow(ctx context.Context, poolID, profileID string) (int64, error)
	ResetShared(ctx context.Context, poolID string, limit int64, ttlSeconds int64) error
	ResetBorrow(ctx context.Context, poolID, profileID string, limit int64, ttlSeconds int64) error
}
