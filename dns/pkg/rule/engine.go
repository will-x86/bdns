package rule

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
)

type Verdict int

const (
	VerdictAllow Verdict = iota
	VerdictBlock
	VerdictPassThrough
)

type Decision struct {
	Verdict Verdict
	Reason  string
}

func Allowed(reason string) Decision { return Decision{VerdictAllow, reason} }
func Blocked(reason string) Decision { return Decision{VerdictBlock, reason} }
func PassThrough() Decision          { return Decision{Verdict: VerdictPassThrough} }

type Rule interface {
	Name() string
	Evaluate(ctx context.Context, rctx *RuleContext) (Decision, error)
}

type Engine struct {
	rules []Rule
}

func NewEngine(rules ...Rule) *Engine {
	return &Engine{rules: rules}
}

func (e *Engine) Evaluate(ctx context.Context, rctx *RuleContext) (Decision, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "rule-engine").Logger()
	ctx = log.WithContext(ctx)

	for _, rule := range e.rules {
		log.Trace().Str("rule", rule.Name()).Msg("running rule")
		d, err := rule.Evaluate(ctx, rctx)
		if err != nil {
			log.Trace().Err(err).Str("reason", d.Reason).Msg("error on rule")
			return Decision{}, fmt.Errorf("rule %q: %w", rule.Name(), err)
		}
		if d.Verdict != VerdictPassThrough {
			log.Trace().Str("rule", rule.Name()).Any("verdict", d.Verdict).Str("reason", d.Reason).Msg("verdict is not pass through")
			return d, nil
		}
	}
	log.Trace().Msg("no rules matched, passing through")
	return Allowed("no rules matched"), nil
}
