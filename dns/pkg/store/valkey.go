package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"github.com/rs/zerolog"
	"github.com/valkey-io/valkey-go"
)

type PoolValkeyStore struct {
	client        valkey.Client
	profilepoolid map[string]string // profile_id : pool_id
}

func NewValkey(addr string) (*PoolValkeyStore, error) {
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
	friendPools, err := db.GetAllFriendPools()
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	err = initValkey(poolValkeyStore.client, friendPools, poolMembers)
	if err != nil {
		return &PoolValkeyStore{}, err
	}
	return &poolValkeyStore, nil

}
func initValkey(client valkey.Client, friendPools []models.FriendPool, poolMembers []models.FriendPoolMembers) error {
	p := PoolValkeyStore{}
	poolIDtoProfiles := make(map[string][]string) // poolID -> []profileIDs
	poolIDtoMode := make(map[string]string)       // poolID -> shared/borrow
	poolIDtoLimit := make(map[string]int64)       // poolID -> TotalLimit

	for _, v := range friendPools {
		poolIDtoProfiles[v.ID] = []string{}
		poolIDtoMode[v.ID] = v.PoolMode
		poolIDtoLimit[v.ID] = v.TotalLimit
	}
	for _, v := range poolMembers {
		poolIDtoProfiles[v.PoolID] = append(poolIDtoProfiles[v.PoolID], v.ProfileID)
	}

	// Validate pool membership constraints before writing anything
	for k, v := range poolIDtoProfiles {
		switch poolIDtoMode[k] {
		case "shared":
			if len(v) < 1 {
				return fmt.Errorf("shared pool %s has no members", k)
			}
		case "borrow":
			if len(v) < 1 {
				return fmt.Errorf("borrow pool %s has no members", k)
			}
		default:
			panic("unavailable pool mode: " + poolIDtoMode[k])
		}
	}

	ctx := context.Background()

	for _, v := range friendPools {
		limit := strconv.FormatInt(v.TotalLimit, 10)
		fmt.Printf("limit: %s %d", limit, v.TotalLimit)
		switch v.PoolMode {
		case "shared":
			// One credit key for the whole pool
			key := p.SharedPoolKey(v.ID)
			resp := client.Do(ctx, client.B().Set().Key(key).Value(limit).Nx().Build())
			if resp.Error() != nil {
				if !valkey.IsValkeyNil(resp.Error()) {
					return fmt.Errorf("failed to set shared pool key %s: %w", key, resp.Error())
				}
			}

		case "borrow":
			// One credit key per member in the pool
			for _, profileID := range poolIDtoProfiles[v.ID] {
				key := p.BorrowPoolKey(v.ID, profileID)
				resp := client.Do(ctx, client.B().Set().Key(key).Value(limit).Nx().Build())
				if resp.Error() != nil {
					if !valkey.IsValkeyNil(resp.Error()) {
						return fmt.Errorf("failed to set borrow pool key %s: %w", key, resp.Error())
					}
				}
			}

		default:
			panic("unavailable pool mode: " + v.PoolMode)
		}
	}

	return nil
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
	log.Debug().Str("key", p.BorrowPoolKey(poolID, profileID)).Msg("decrementing remaining borrow for ")
	return p.client.Do(ctx, p.client.B().Decr().Key(p.BorrowPoolKey(poolID, profileID)).Build()).Error()
}
func (p *PoolValkeyStore) DecrementRemainingShared(ctx context.Context, poolID string) error {
	log := zerolog.Ctx(ctx).With().Str("component", "valkey-store").Logger()
	ctx = log.WithContext(ctx)
	log.Debug().Str("key", p.SharedPoolKey(poolID)).Msg("decrementing remaining shared for ")
	return p.client.Do(ctx, p.client.B().Decr().Key(p.SharedPoolKey(poolID)).Build()).Error()
}
