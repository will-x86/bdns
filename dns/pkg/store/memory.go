package store

import (
	"context"
	"errors"
)

type PoolMemory struct {
	pools map[string]string
}

func (p *PoolMemory) Exists(ctx context.Context, profileID, poolID string) bool {
	return false
}

func (p *PoolMemory) PoolID(ctx context.Context, profileID string) (string, error) {
	return "", errors.New("unimplemented")
}
func NewMemory() *PoolMemory {
	return &PoolMemory{}

}
