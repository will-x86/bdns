package proxy

import (
	"codeberg.org/will-x86/bdns/dns/pkg/rule"
	"codeberg.org/will-x86/bdns/dns/pkg/rule/rules"
)

// all active rules in *priority* order
func BuildEngine(stores rule.Stores) *rule.Engine {
	return rule.NewEngine(
		&rules.PermanentWhitelistRule{},
		&rules.TemporaryWhitelistRule{},
		&rules.CategoryBlockRule{},
		&rules.TimeBlockRule{},
		&rules.SharedPoolRule{},
		&rules.BorrowPoolRule{},
	)
}
