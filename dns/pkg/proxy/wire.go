package proxy

import (
	"github.com/will-x86/bdns/dns/pkg/rule"
	"github.com/will-x86/bdns/dns/pkg/rule/rules"
)

// all active rules in *priority* order
func BuildEngine(stores rule.Stores) *rule.Engine {
	return rule.NewEngine(
		&rules.PermanentWhitelistRule{},
		&rules.TemporaryWhitelistRule{},
		&rules.CategoryBlockRule{},
	)
}
