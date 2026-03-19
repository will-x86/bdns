package rules

import (
	"context"

	"codeberg.org/will-x86/bdns/dns/pkg/rule"
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

	return rule.Allowed("ahh"), nil
}
