package rules

import (
	"context"
	"log"

	"github.com/will-x86/bdns/dns/pkg/rule"
)

type TimeBlockRule struct{}

func (r *TimeBlockRule) Name() string { return "permanent_whitelist" }

func (r *TimeBlockRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	// is this category currently fully blocked for the user
	// e.g. "block social media 9-12am
	// factors into account users timezzone
	category, err := rctx.GetCategory(ctx)
	if err != nil {
		return rule.Decision{}, err

	}
	// No category, therefore cannot have timeblock block
	if category == "" {
		return rule.PassThrough(), nil
	}
	blocks, err := rctx.Stores.TimeBlock.GetTimeBlocks(ctx, rctx.ProfileID, category)
	if err != nil {
		return rule.Decision{}, err
	}
	for k, v := range blocks {
		log.Printf("debug print block: %d, %+v", k, v)
	}
	return rule.PassThrough(), nil
}
