package rules

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type SharedPoolRule struct{}

func (r *SharedPoolRule) Name() string { return "shared_pool_block" }

func (r *SharedPoolRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	// hit PoolCache to see if rctx.Stores.PoolCache.PoolExists(profile_id)
	log := zerolog.Ctx(ctx).With().Str("component", "shared-pool").Logger()
	ctx = log.WithContext(ctx)

	log.Trace().Str("profile_id", rctx.ProfileID).Msg("entering evaluate on shared pool rule")
	poolID, err := rctx.Stores.PoolCache.PoolID(ctx, rctx.ProfileID)
	if err != nil {
		log.Trace().Msg("passing through as no poolID")
		return rule.PassThrough(), nil
	}

	log = zerolog.Ctx(ctx).With().Str("poolID", poolID).Logger()
	ctx = log.WithContext(ctx)

	if !rctx.Stores.PoolCache.ExistsShared(ctx, poolID) {
		log.Trace().Msg("passing through as no shared pool key in cache")
		return rule.PassThrough(), nil
	}
	pool, err := rctx.GetPool(ctx, poolID)
	//pool, err := rctx.Stores.PoolDB.GetPool(ctx, poolID)
	if err != nil {
		log.Debug().Msg("pool exists but cannot get")
		return rule.Decision{}, err
	}

	if pool.PoolMode == "borrow" {
		log.Trace().Msg("pool mode is borrow, in shared, passing through")
		return rule.PassThrough(), nil
	} else if pool.PoolMode != "shared" {
		log.Trace().Msg("pool mode is not shared or borrow in shared, passing through")
		return rule.Decision{}, fmt.Errorf("rule is neither shared or borrow, rule is %s", pool.PoolMode)
	}
	category, err := rctx.GetCategory(ctx)
	if err != nil {
		log.Trace().Msg("error on get category, returning error")
		return rule.Decision{}, err
	}
	log = zerolog.Ctx(ctx).With().Str("category", category).Logger()
	ctx = log.WithContext(ctx)

	if !rctx.Stores.PoolDB.PoolCategoryBlocked(ctx, poolID, category) {
		log.Trace().Msg("category not blocked, passing through")
		return rule.PassThrough(), nil
	}
	remaining, err := rctx.Stores.PoolCache.GetRemainingShared(ctx, poolID)
	if err != nil {
		return rule.Decision{}, err
	}
	log.Trace().Int64("remaining_limit", remaining).Send()
	if remaining > 0 {
		err = rctx.Stores.PoolCache.DecrementRemainingShared(ctx, poolID)
		if err != nil {
			return rule.Decision{}, err
		}
		//return rule.Allowed("some remaining on this shared pool"), nil
		return rule.PassThrough(), nil

	} else {
		return rule.Blocked("no more remaining on this shared pool"), nil
	}

}
