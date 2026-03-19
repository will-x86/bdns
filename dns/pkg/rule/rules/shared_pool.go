package rules

import (
	"context"
	"fmt"

	"codeberg.org/will-x86/bdns/dns/pkg/rule"
	"github.com/rs/zerolog"
)

type SharedPoolRule struct{}

func (r *SharedPoolRule) Name() string { return "shared_pool_block" }

func (r *SharedPoolRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	// hit PoolCache to see if rctx.Stores.PoolCache.PoolExists(profile_id)
	log := zerolog.Ctx(ctx).With().Str("component", "shared-pool").Logger()
	ctx = log.WithContext(ctx)
	poolID, err := rctx.Stores.PoolCache.PoolID(ctx, rctx.ProfileID)
	if err != nil {
		return rule.PassThrough(), nil
	}
	if !rctx.Stores.PoolCache.Exists(ctx, rctx.ProfileID, poolID) {
		// no pool exists
		return rule.PassThrough(), nil
	}
	pool, err := rctx.GetPool(ctx, poolID)
	//pool, err := rctx.Stores.PoolDB.GetPool(ctx, poolID)
	if err != nil {
		return rule.Decision{}, err
	}

	if pool.PoolMode == "borrow" {
		return rule.PassThrough(), nil
	} else if pool.PoolMode != "shared" {
		return rule.Decision{}, fmt.Errorf("rule is neither shared or borrow, rule is %s", pool.PoolMode)
	}
	category, err := rctx.GetCategory(ctx)
	if err != nil {
		return rule.Decision{}, err
	}
	if !rctx.Stores.PoolDB.PoolCategoryBlocked(ctx, poolID, category) {
		return rule.PassThrough(), nil
	}
	// Domain is blocked in that cateogory so we do uh
	// rctx.Stores.PoolCache.DecrementLimit(pool_id) -> if err (no limit remaining), return ...
	// To blame on users, we'll do that later
	return rule.Allowed("ahh"), nil
}
