package store

import "context"

type Pool interface {
	PoolID(ctx context.Context, profileID string) (string, error)
	ExistsShared(ctx context.Context, poolID string) bool
	ExistsBorrow(ctx context.Context, poolID, profileID string) bool
	GetRemainingShared(ctx context.Context, poolID string) (int64, error)
	DecrementRemainingBorrow(ctx context.Context, poolID, profileID string) error
	DecrementRemainingShared(ctx context.Context, poolID string) error
	GetRemainingBorrow(ctx context.Context, poolID, profileID string) (int64, error)
	ResetShared(ctx context.Context, poolID string, limit int64, ttlSeconds int64) error
	ResetBorrow(ctx context.Context, poolID, profileID string, limit int64, ttlSeconds int64) error
}
