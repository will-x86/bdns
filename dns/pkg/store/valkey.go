package store

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/valkey-io/valkey-go"
)

type PoolValkeyStore struct {
	client        valkey.Client
	pools         map[string]string // string to pool...
	profilepoolid map[string]string // profile_id : pool_id
}

func NewValkey(addr string) (*PoolValkeyStore, error) {
	var err error
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	return &PoolValkeyStore{
		client: client,
	}, nil

}

// Returns pool_id from profile_id
func (p *PoolValkeyStore) PoolID(ctx context.Context, profileID string) (string, error) {
	if v, ok := p.profilepoolid[profileID]; !ok {
		return "", errors.New("pool does not exist with given profile_id")
	} else {
		return v, nil
	}
}
func (p *PoolValkeyStore) SharedPoolKey(poolID string) string {
	return "pool:" + poolID + ":credits"
}
func (p *PoolValkeyStore) BorrowPoolKey(poolID, profileID string) string {
	return "pool:" + poolID + ":" + profileID + ":credits"
}
func (p *PoolValkeyStore) Exists(ctx context.Context, profileID, poolID string) bool {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)

	// check valkey if pool:{pool_id}:credits = total_limit exists
	// or pool:{pool_id}:{profile_id}:credits = total_limit
	//	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	resp := p.client.Do(ctx, p.client.B().Exists().Key(p.SharedPoolKey(poolID), p.BorrowPoolKey(poolID, profileID)).Build())
	numExist, err := resp.AsInt64()
	if err != nil {
		log.Warn().Err(err).Msg("Valkey exists, error turning 'exists' into int64")
		return false
	}
	return numExist != 0
}
