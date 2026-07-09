package rules

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type BorrowPoolRule struct{}

func (r *BorrowPoolRule) Name() string { return "borrow_pool_block" }

func (r *BorrowPoolRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	// Check if rctx.Pool or whatever ( nil pointer)
	// If it does, see rctx.PoolMode, should be borrow
	// See if category is banned via rctx.Stores.Resolve.ResolveCategory()
	// Then check rctx.Stores.PoolDB.CateogryBlocked(pool_id) which checks friend_pool_category_blocks
	// If it's blocked:
	// rctx.Stores.PoolCache.DecrementLimit(pool_id) -> if err (no limit remaining), return ...
	// To blame on users, we'll do that later
	// hit PoolCache to see if rctx.Stores.PoolCache.PoolExists(profile_id)
	log := zerolog.Ctx(ctx).With().Str("component", "borrow-pool").Str("profile_id", rctx.ProfileID).Logger()
	ctx = log.WithContext(ctx)
	log.Trace().Msg("entering evaluate on borrow pool rule")

	poolID, err := rctx.Stores.PoolCache.PoolID(ctx, rctx.ProfileID)
	if err != nil {
		log.Trace().Msg("passing through as no poolID")
		return rule.PassThrough(), nil
	}

	log = log.With().Str("pool_id", poolID).Logger()
	ctx = log.WithContext(ctx)

	if !rctx.Stores.PoolCache.ExistsBorrow(ctx, poolID, rctx.ProfileID) {
		log.Trace().Msg("passing through as no borrow pool key in cache")
		return rule.PassThrough(), nil
	}

	pool, err := rctx.GetPool(ctx, poolID)
	//pool, err := rctx.Stores.PoolDB.GetPool(ctx, poolID)
	if err != nil {
		log.Trace().Msg("passing through as no pool in cache")
		return rule.Decision{}, err
	}

	log = log.With().Str("pool_mode", pool.PoolMode).Logger()
	ctx = log.WithContext(ctx)

	if pool.PoolMode == "shared" {
		log.Trace().Msg("pool mode is shared, in borrow, passing through")
		return rule.PassThrough(), nil
	} else if pool.PoolMode != "borrow" {
		log.Trace().Msg("pool mode is not shared or borrow in borrow, passing through")
		return rule.Decision{}, fmt.Errorf("rule is neither shared or borrow, rule is %s", pool.PoolMode)
	}

	category, err := rctx.GetCategory(ctx)
	if err != nil {
		log.Trace().Msg("pool mode is not shared or borrow in shared, passing through")
		return rule.Decision{}, err
	}

	log = log.With().Str("category", category).Logger()
	ctx = log.WithContext(ctx)

	if !rctx.Stores.PoolDB.PoolCategoryBlocked(ctx, poolID, category) {
		log.Trace().Msg("category not blocked, passing through")
		return rule.PassThrough(), nil
	}

	remaining, err := rctx.Stores.PoolCache.GetRemainingBorrow(ctx, poolID, rctx.ProfileID)
	if err != nil {
		return rule.Decision{}, err
	}

	log.Trace().Int64("remaining_limit", remaining).Send()

	if remaining > 0 {
		err = rctx.Stores.PoolCache.DecrementRemainingBorrow(ctx, poolID, rctx.ProfileID)
		if err != nil {
			return rule.Decision{}, err
		}
		//return rule.Allowed("some remaining on this borrow pool"), nil
		return rule.PassThrough(), nil
	} else {
		return rule.Blocked("no more remaining on this borrow pool"), nil
	}
}
