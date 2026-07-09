package proxy

import (
	"testing"

	"github.com/will-x86/bdns/dns/pkg/rule"
)

func TestBuildEngine(t *testing.T) {
	if e := BuildEngine(rule.Stores{}); e == nil {
		t.Fatal("BuildEngine returned nil")
	}
}
