package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"github.com/rs/zerolog"
	"github.com/valkey-io/valkey-go"
)

type PoolValkeyStore struct {
	client        valkey.Client
	profilepoolid map[string]string // profile_id : pool_id
}

func NewValkey(ctx context.Context, addr string, stores *db.SQLiteStores) (*PoolValkeyStore, error) {
	var err error
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	poolMembers, err := db.GetAllFriendPoolMembers()
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	poolValkeyStore := PoolValkeyStore{
		client:        client,
		profilepoolid: make(map[string]string),
	}
	for k := range poolMembers {
		poolValkeyStore.profilepoolid[poolMembers[k].ProfileID] = poolMembers[k].PoolID
	}
	allMembers, err := stores.GetAllPoolMembersWithTimezones(ctx)
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	err = initValkey(ctx, client, allMembers, time.Now())
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	return &poolValkeyStore, nil

}
func initValkey(ctx context.Context, client valkey.Client, allMembers []models.AllMembers, now time.Time) error {
	p := PoolValkeyStore{}
	for _, v := range allMembers {
		// Shared we're creating the keys based on creator
		secondsTil4am := SecondsUntil4AM(ctx, v.Timezone, now)
		if v.PoolMode == "shared" && v.CreatedBy == v.UserID {
			key := p.SharedPoolKey(v.PoolID)
			limit := strconv.FormatInt(v.TotalLimit, 10)
			resp := client.Do(ctx, client.B().Set().Key(key).Value(limit).Nx().ExSeconds(int64(secondsTil4am)).Build())
			if resp.Error() != nil {
				if !valkey.IsValkeyNil(resp.Error()) {
					return fmt.Errorf("failed to set shared pool key %s: %w", key, resp.Error())
				}
			}
		} else if v.PoolMode == "borrow" {
			key := p.BorrowPoolKey(v.PoolID, v.ProfileID)
			limit := strconv.FormatInt(v.TotalLimit, 10)
			resp := client.Do(ctx, client.B().Set().Key(key).Value(limit).Nx().ExSeconds(int64(secondsTil4am)).Build())
			if resp.Error() != nil {
				if !valkey.IsValkeyNil(resp.Error()) {
					return fmt.Errorf("failed to set borrow pool key %s: %w", key, resp.Error())
				}
			}

		}
	}

	return nil
}

// Returns pool_id from profile_id
func (p *PoolValkeyStore) PoolID(ctx context.Context, profileID string) (string, error) {
	v, ok := p.profilepoolid[profileID]
	if !ok {
		return "", errors.New("pool does not exist with given profile_id")
	}
	return v, nil
}

func (p *PoolValkeyStore) SharedPoolKey(poolID string) string {
	return "pool:" + poolID + ":credits"
}
func (p *PoolValkeyStore) BorrowPoolKey(poolID, profileID string) string {
	return "pool:" + poolID + ":" + profileID + ":credits"
}

func (p *PoolValkeyStore) ExistsShared(ctx context.Context, poolID string) bool {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	resp := p.client.Do(ctx, p.client.B().Exists().Key(p.SharedPoolKey(poolID)).Build())
	n, err := resp.AsInt64()
	if err != nil {
		log.Warn().Err(err).Msg("valkey ExistsShared error")
		return false
	}
	return n != 0
}

func (p *PoolValkeyStore) ExistsBorrow(ctx context.Context, poolID, profileID string) bool {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	resp := p.client.Do(ctx, p.client.B().Exists().Key(p.BorrowPoolKey(poolID, profileID)).Build())
	n, err := resp.AsInt64()
	if err != nil {
		log.Warn().Err(err).Msg("valkey ExistsBorrow error")
		return false
	}
	return n != 0
}

func (p *PoolValkeyStore) GetRemainingShared(ctx context.Context, poolID string) (int64, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)
	log.Debug().Str("key", p.SharedPoolKey(poolID)).Msg("get remaining shared for ")
	resp := p.client.Do(ctx, p.client.B().Get().Key(p.SharedPoolKey(poolID)).Build())
	if resp.Error() != nil {
		return 0, resp.Error()
	}
	remain, err := resp.AsInt64()
	if err != nil {
		return 0, err
	}

	return remain, nil

}

func (p *PoolValkeyStore) GetRemainingBorrow(ctx context.Context, poolID, profileID string) (int64, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)
	log.Debug().Str("key", p.BorrowPoolKey(poolID, profileID)).Msg("get remaining borrow for ")
	resp := p.client.Do(ctx, p.client.B().Get().Key(p.BorrowPoolKey(poolID, profileID)).Build())
	if resp.Error() != nil {
		return 0, resp.Error()
	}
	remain, err := resp.AsInt64()
	if err != nil {
		return 0, err
	}

	return remain, nil
}

func (p *PoolValkeyStore) DecrementRemainingBorrow(ctx context.Context, poolID, profileID string) error {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)
	key := p.BorrowPoolKey(poolID, profileID)
	log.Debug().Str("key", key).Msg("decrementing remaining borrow for ")

	ttl, err := p.client.Do(ctx, p.client.B().Ttl().Key(key).Build()).AsInt64()
	if err != nil {
		return err
	}

	if err := p.client.Do(ctx, p.client.B().Decr().Key(key).Build()).Error(); err != nil {
		return err
	}

	if ttl > 0 {
		return p.client.Do(ctx, p.client.B().Expire().Key(key).Seconds(ttl).Build()).Error()
	}

	return nil
}
func (p *PoolValkeyStore) DecrementRemainingShared(ctx context.Context, poolID string) error {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)
	key := p.SharedPoolKey(poolID)
	log.Debug().Str("key", key).Msg("decrementing remaining borrow for ")

	ttl, err := p.client.Do(ctx, p.client.B().Ttl().Key(key).Build()).AsInt64()
	if err != nil {
		return err
	}

	if err := p.client.Do(ctx, p.client.B().Decr().Key(key).Build()).Error(); err != nil {
		return err
	}

	if ttl > 0 {
		return p.client.Do(ctx, p.client.B().Expire().Key(key).Seconds(ttl).Build()).Error()
	}

	return nil
}

func (p *PoolValkeyStore) ResetShared(ctx context.Context, poolID string, limit int64, ttlSeconds int64) error {
	// Pool should exist,
	key := p.SharedPoolKey(poolID)
	limitS := strconv.FormatInt(limit, 10)
	resp := p.client.Do(ctx, p.client.B().Set().Key(key).Value(limitS).ExSeconds(ttlSeconds).Build())
	return resp.Error()

}

func (p *PoolValkeyStore) ResetBorrow(ctx context.Context, poolID, profileID string, limit int64, ttlSeconds int64) error {
	// Pool should exist,
	key := p.BorrowPoolKey(poolID, profileID)
	limitS := strconv.FormatInt(limit, 10)
	resp := p.client.Do(ctx, p.client.B().Set().Key(key).Value(limitS).ExSeconds(ttlSeconds).Build())
	return resp.Error()

}
