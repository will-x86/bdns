package rules

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type PermanentWhitelistRule struct{}

func (r *PermanentWhitelistRule) Name() string { return "permanent_whitelist" }

func (r *PermanentWhitelistRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "permanent-whitelist-rule").Logger()
	ctx = log.WithContext(ctx)
	log.Trace().Msg("entering evaluate")
	// is domain permanently whitelisted for said user
	ok, err := rctx.Stores.Whitelist.IsPermanentlyWhitelisted(ctx, rctx.ProfileID, rctx.Domain)
	if err != nil {
		return rule.Decision{}, err
	}
	if ok {
		return rule.Allowed("permanent whitelist"), nil
	}
	return rule.PassThrough(), nil
}

type TemporaryWhitelistRule struct{}

func (r *TemporaryWhitelistRule) Name() string { return "temporary_whitelist" }

func (r *TemporaryWhitelistRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "temp-whitelist-rule").Logger()
	ctx = log.WithContext(ctx)
	log.Trace().Msg("entering evaluate")
	// is domain temp whitelisted for said user
	ok, err := rctx.Stores.Whitelist.IsTemporarilyWhitelisted(ctx, rctx.ProfileID, rctx.Domain, rctx.Now)
	if err != nil {
		return rule.Decision{}, err
	}
	if ok {
		return rule.Allowed("temporary whitelist"), nil
	}
	return rule.PassThrough(), nil
}
