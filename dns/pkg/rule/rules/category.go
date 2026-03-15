package rules

import (
	"context"

	"codeberg.org/will-x86/bdns/dns/pkg/rule"
	"github.com/rs/zerolog"
)

type CategoryBlockRule struct{}

func (r *CategoryBlockRule) Name() string { return "category_block" }

func (r *CategoryBlockRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "category-rule").Logger()
	ctx = log.WithContext(ctx)
	log.Trace().Msg("entering evaluate")
	// First rule that needs category, so cache in rctx
	category, err := rctx.GetCategory(ctx)
	if err != nil {
		return rule.Decision{}, err
	}

	// Domain not in any blocklist — nothing to check, let it through.
	if category == "" {
		return rule.PassThrough(), nil
	}

	blocked, err := rctx.Stores.Category.IsCategoryBlocked(ctx, rctx.ProfileID, category)
	if err != nil {
		return rule.Decision{}, err
	}
	if blocked {
		return rule.Blocked("category blocked: " + category), nil
	}
	return rule.PassThrough(), nil
}
