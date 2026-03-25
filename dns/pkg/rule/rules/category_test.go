package rules

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"codeberg.org/will-x86/bdns/dns/pkg/rule"
)

// ---------------------------------------------------------------------------
// Fake category store
// ---------------------------------------------------------------------------

type fakeCategoryStore struct {
	blocked bool
	err     error
}

func (f fakeCategoryStore) IsCategoryBlocked(_ context.Context, _, _ string) (bool, error) {
	return f.blocked, f.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeCategoryRuleCtx(category string, catStore fakeCategoryStore, resolveErr error) *rule.RuleContext {
	return &rule.RuleContext{
		Domain:    "example.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			Category: catStore,
			Resolve: func(_ context.Context, _ string) (string, error) {
				return category, resolveErr
			},
		},
	}
}

// ---------------------------------------------------------------------------
// CategoryBlockRule tests
// ---------------------------------------------------------------------------

func TestCategoryBlockRule_Name(t *testing.T) {
	if (&CategoryBlockRule{}).Name() != "category_block" {
		t.Fatal("wrong name")
	}
}

func TestCategoryBlockRule_NoCategory_PassThrough(t *testing.T) {
	// Domain not in any blocklist — category resolves to ""
	rctx := makeCategoryRuleCtx("", fakeCategoryStore{blocked: true}, nil)
	d, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough for uncategorised domain, got %v", d.Verdict)
	}
}

func TestCategoryBlockRule_CategoryBlocked_Blocked(t *testing.T) {
	rctx := makeCategoryRuleCtx("porn", fakeCategoryStore{blocked: true}, nil)
	d, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block for blocked category, got %v", d.Verdict)
	}
}

func TestCategoryBlockRule_CategoryNotBlocked_PassThrough(t *testing.T) {
	rctx := makeCategoryRuleCtx("social", fakeCategoryStore{blocked: false}, nil)
	d, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough for unblocked category, got %v", d.Verdict)
	}
}

func TestCategoryBlockRule_ResolveError_ReturnsError(t *testing.T) {
	resolveErr := errors.New("resolve failure")
	rctx := makeCategoryRuleCtx("", fakeCategoryStore{}, resolveErr)
	_, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, resolveErr) {
		t.Fatalf("want resolve error, got %v", err)
	}
}

func TestCategoryBlockRule_StoreError_ReturnsError(t *testing.T) {
	storeErr := errors.New("db failure")
	rctx := makeCategoryRuleCtx("social", fakeCategoryStore{err: storeErr}, nil)
	_, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, storeErr) {
		t.Fatalf("want store error, got %v", err)
	}
}

func TestCategoryBlockRule_ReasonContainsCategory(t *testing.T) {
	rctx := makeCategoryRuleCtx("gambling", fakeCategoryStore{blocked: true}, nil)
	d, err := (&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictBlock {
		t.Fatalf("want Block, got %v", d.Verdict)
	}
	if d.Reason == "" {
		t.Fatal("want non-empty reason")
	}
	want := "gambling"
	if !containsStr(d.Reason, want) {
		t.Fatalf("reason %q does not contain category %q", d.Reason, want)
	}
}

// Category is cached in RuleContext — resolver should only be called once
func TestCategoryBlockRule_CategoryCachedInRuleContext(t *testing.T) {
	calls := 0
	rctx := &rule.RuleContext{
		Domain:    "example.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			Category: fakeCategoryStore{blocked: false},
			Resolve: func(_ context.Context, _ string) (string, error) {
				calls++
				return "social", nil
			},
		},
	}
	(&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	(&CategoryBlockRule{}).Evaluate(context.Background(), rctx)
	if calls > 1 {
		t.Fatalf("Resolve should be called at most once due to RuleContext cache, got %d calls", calls)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
