package store

import (
	"context"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	"github.com/rs/zerolog"
)

type Resetter struct {
	stores *db.SQLiteStores
	pool   Pool
}

func NewResetter(stores *db.SQLiteStores, pool Pool) Resetter {
	return Resetter{stores: stores, pool: pool}
}
func (r *Resetter) StartResetJob(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	log := zerolog.Ctx(ctx).With().Str("component", "resetter").Logger()
	log.Trace().Msg("starting reset ticker")
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Trace().Msg("starting ticker reset now")
				r.resetPools(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}
func (r *Resetter) resetPools(ctx context.Context) {
	log := zerolog.Ctx(ctx).With().Str("component", "resetter").Logger()
	members, err := r.stores.GetAllPoolMembersWithTimezones(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting all pool members with timezones")
	}
	now := time.Now()
	for _, v := range members {
		log.Trace().Any("member", v).Send()
		ttl := int64(SecondsUntil4AM(ctx, v.Timezone, now))
		if v.PoolMode == "shared" && v.CreatedBy == v.UserID {
			// Doesn't exist, TTL expired it
			if !r.pool.Exists(ctx, v.ProfileID, v.PoolID) {
				err := r.pool.ResetShared(ctx, v.PoolID, v.TotalLimit, ttl)
				if err != nil {
					log.Fatal().Err(err).Msg("error resetting shared pool")
				}
				log.Trace().Str("poolID", v.PoolID).Str("profileID", v.ProfileID).Msg("reset shared for pool")
			} else {
				log.Trace().Str("poolID", v.PoolID).Msg("skipping reset on pool as exists")
			}
		} else if v.PoolMode == "borrow" {
			// Doesn't exist, TTL expired it
			if !r.pool.Exists(ctx, v.ProfileID, v.PoolID) {
				err := r.pool.ResetBorrow(ctx, v.PoolID, v.ProfileID, v.TotalLimit, ttl)
				if err != nil {
					log.Fatal().Err(err).Msg("error resetting shared pool")
				}
				log.Trace().Str("poolID", v.PoolID).Str("profileID", v.ProfileID).Msg("reset borrow for pool")
			} else {
				log.Trace().Str("poolID", v.PoolID).Msg("skipping reset on pool as exists")
			}
		}
	}

}
