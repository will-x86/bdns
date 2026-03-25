package rules

import (
	"context"
	"errors"
	"testing"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"codeberg.org/will-x86/bdns/dns/pkg/rule"
)

// Fake whitelist store

type fakeWhitelistStore struct {
	permanent    bool
	temporary    bool
	permanentErr error
	temporaryErr error
}

func (f fakeWhitelistStore) IsPermanentlyWhitelisted(_ context.Context, _, _ string) (bool, error) {
	return f.permanent, f.permanentErr
}
func (f fakeWhitelistStore) IsTemporarilyWhitelisted(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	return f.temporary, f.temporaryErr
}

func makeWhitelistRuleCtx(domain string, store fakeWhitelistStore) *rule.RuleContext {
	return &rule.RuleContext{
		Domain:    domain,
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores:    rule.Stores{Whitelist: store},
	}
}

// PermanentWhitelistRule tests

func TestPermanentWhitelistRule_Name(t *testing.T) {
	if (&PermanentWhitelistRule{}).Name() != "permanent_whitelist" {
		t.Fatal("wrong name")
	}
}

func TestPermanentWhitelistRule_Whitelisted_Allowed(t *testing.T) {
	rctx := makeWhitelistRuleCtx("ads.example.com", fakeWhitelistStore{permanent: true})
	d, err := (&PermanentWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow for whitelisted domain, got %v", d.Verdict)
	}
}

func TestPermanentWhitelistRule_NotWhitelisted_PassThrough(t *testing.T) {
	rctx := makeWhitelistRuleCtx("evil.com", fakeWhitelistStore{permanent: false})
	d, err := (&PermanentWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough for non-whitelisted domain, got %v", d.Verdict)
	}
}

func TestPermanentWhitelistRule_StoreError_ReturnsError(t *testing.T) {
	storeErr := errors.New("db failure")
	rctx := makeWhitelistRuleCtx("evil.com", fakeWhitelistStore{permanentErr: storeErr})
	_, err := (&PermanentWhitelistRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, storeErr) {
		t.Fatalf("want store error, got %v", err)
	}
}

func TestPermanentWhitelistRule_ReasonSet(t *testing.T) {
	rctx := makeWhitelistRuleCtx("safe.com", fakeWhitelistStore{permanent: true})
	d, err := (&PermanentWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Reason == "" {
		t.Fatal("want non-empty reason on Allow")
	}
}

// TemporaryWhitelistRule tests

func TestTemporaryWhitelistRule_Name(t *testing.T) {
	if (&TemporaryWhitelistRule{}).Name() != "temporary_whitelist" {
		t.Fatal("wrong name")
	}
}

func TestTemporaryWhitelistRule_Whitelisted_Allowed(t *testing.T) {
	rctx := makeWhitelistRuleCtx("tracker.example.com", fakeWhitelistStore{temporary: true})
	d, err := (&TemporaryWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow for temp-whitelisted domain, got %v", d.Verdict)
	}
}

func TestTemporaryWhitelistRule_NotWhitelisted_PassThrough(t *testing.T) {
	rctx := makeWhitelistRuleCtx("evil.com", fakeWhitelistStore{temporary: false})
	d, err := (&TemporaryWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictPassThrough {
		t.Fatalf("want PassThrough for non-temp-whitelisted domain, got %v", d.Verdict)
	}
}

func TestTemporaryWhitelistRule_StoreError_ReturnsError(t *testing.T) {
	storeErr := errors.New("db failure")
	rctx := makeWhitelistRuleCtx("evil.com", fakeWhitelistStore{temporaryErr: storeErr})
	_, err := (&TemporaryWhitelistRule{}).Evaluate(context.Background(), rctx)
	if !errors.Is(err, storeErr) {
		t.Fatalf("want store error, got %v", err)
	}
}

func TestTemporaryWhitelistRule_ReasonSet(t *testing.T) {
	rctx := makeWhitelistRuleCtx("temp.com", fakeWhitelistStore{temporary: true})
	d, err := (&TemporaryWhitelistRule{}).Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Reason == "" {
		t.Fatal("want non-empty reason on Allow")
	}
}

// Interaction: permanent whitelist short-circuits before temporary

func TestWhitelistRules_PermanentBeforeTemporaryInEngine(t *testing.T) {
	// Engine with permanent then temporary: if permanent fires, temporary is never reached
	temporaryCalled := false
	eng := rule.NewEngine(
		&PermanentWhitelistRule{},
		&TemporaryWhitelistRule{},
	)
	rctx := &rule.RuleContext{
		Domain:    "safe.com",
		ProfileID: "profile1",
		Now:       time.Now(),
		User:      &models.User{Timezone: "UTC"},
		Stores: rule.Stores{
			Whitelist: &trackingWhitelistStore{
				permanent: true,
				onTemporary: func() {
					temporaryCalled = true
				},
			},
		},
	}
	d, err := eng.Evaluate(context.Background(), rctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != rule.VerdictAllow {
		t.Fatalf("want Allow from permanent whitelist, got %v", d.Verdict)
	}
	if temporaryCalled {
		t.Fatal("temporary whitelist store should not be called when permanent fires first")
	}
}

type trackingWhitelistStore struct {
	permanent   bool
	onTemporary func()
}

func (f *trackingWhitelistStore) IsPermanentlyWhitelisted(_ context.Context, _, _ string) (bool, error) {
	return f.permanent, nil
}
func (f *trackingWhitelistStore) IsTemporarilyWhitelisted(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	if f.onTemporary != nil {
		f.onTemporary()
	}
	return false, nil
}
