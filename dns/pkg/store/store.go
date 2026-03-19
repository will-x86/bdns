package store

import "context"

type Pool interface {
	Exists(ctx context.Context, profileID, poolID string) bool
	PoolID(ctx context.Context, profileID string) (string, error)
}
