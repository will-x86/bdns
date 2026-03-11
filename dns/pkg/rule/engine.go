package rule

import (
	"context"
	"fmt"
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
	for _, rule := range e.rules {
		d, err := rule.Evaluate(ctx, rctx)
		if err != nil {
			return Decision{}, fmt.Errorf("rule %q: %w", rule.Name(), err)
		}
		if d.Verdict != VerdictPassThrough {
			return d, nil
		}
	}
	return Allowed("no rules matched"), nil
}
