package store

import (
	"context"
	"testing"
)

var ctx = context.Background()

// PoolID

func TestMemory_PoolID_NotFound(t *testing.T) {
	m := NewMemory()
	_, err := m.PoolID(ctx, "unknown-profile")
	if err == nil {
		t.Fatal("want error for unknown profile, got nil")
	}
}

func TestMemory_PoolID_Found(t *testing.T) {
	m := NewMemory()
	m.SetProfilePool("profile1", "pool1")
	got, err := m.PoolID(ctx, "profile1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "pool1" {
		t.Fatalf("want pool1, got %q", got)
	}
}

// ExistsShared / ExistsBorrow

func TestMemory_ExistsShared_Missing(t *testing.T) {
	m := NewMemory()
	if m.ExistsShared(ctx, "pool1") {
		t.Fatal("want false for missing shared key")
	}
}

func TestMemory_ExistsShared_Present(t *testing.T) {
	m := NewMemory()
	m.SetShared("pool1", 100)
	if !m.ExistsShared(ctx, "pool1") {
		t.Fatal("want true after SetShared")
	}
}

func TestMemory_ExistsBorrow_Missing(t *testing.T) {
	m := NewMemory()
	if m.ExistsBorrow(ctx, "pool1", "profile1") {
		t.Fatal("want false for missing borrow key")
	}
}

func TestMemory_ExistsBorrow_Present(t *testing.T) {
	m := NewMemory()
	m.SetBorrow("pool1", "profile1", 50)
	if !m.ExistsBorrow(ctx, "pool1", "profile1") {
		t.Fatal("want true after SetBorrow")
	}
}

// ExistsBorrow is profile-scoped — another profile's key doesn't count
func TestMemory_ExistsBorrow_WrongProfile(t *testing.T) {
	m := NewMemory()
	m.SetBorrow("pool1", "profile1", 50)
	if m.ExistsBorrow(ctx, "pool1", "profile2") {
		t.Fatal("want false for different profile")
	}
}

// GetRemainingShared

func TestMemory_GetRemainingShared_NotFound(t *testing.T) {
	m := NewMemory()
	_, err := m.GetRemainingShared(ctx, "pool1")
	if err == nil {
		t.Fatal("want error for missing shared pool")
	}
}

func TestMemory_GetRemainingShared_ReturnsValue(t *testing.T) {
	m := NewMemory()
	m.SetShared("pool1", 42)
	got, err := m.GetRemainingShared(ctx, "pool1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 42 {
		t.Fatalf("want 42, got %d", got)
	}
}

// GetRemainingBorrow

func TestMemory_GetRemainingBorrow_NotFound(t *testing.T) {
	m := NewMemory()
	_, err := m.GetRemainingBorrow(ctx, "pool1", "profile1")
	if err == nil {
		t.Fatal("want error for missing borrow pool")
	}
}

func TestMemory_GetRemainingBorrow_ReturnsValue(t *testing.T) {
	m := NewMemory()
	m.SetBorrow("pool1", "profile1", 7)
	got, err := m.GetRemainingBorrow(ctx, "pool1", "profile1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Fatalf("want 7, got %d", got)
	}
}

// DecrementRemainingShared

func TestMemory_DecrementShared_NotFound(t *testing.T) {
	m := NewMemory()
	err := m.DecrementRemainingShared(ctx, "pool1")
	if err == nil {
		t.Fatal("want error decrementing missing shared pool")
	}
}

func TestMemory_DecrementShared_Decrements(t *testing.T) {
	m := NewMemory()
	m.SetShared("pool1", 10)
	if err := m.DecrementRemainingShared(ctx, "pool1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := m.GetRemainingShared(ctx, "pool1")
	if got != 9 {
		t.Fatalf("want 9 after decrement, got %d", got)
	}
}

func TestMemory_DecrementShared_MultipleDecrements(t *testing.T) {
	m := NewMemory()
	m.SetShared("pool1", 3)
	for i := 0; i < 3; i++ {
		if err := m.DecrementRemainingShared(ctx, "pool1"); err != nil {
			t.Fatalf("decrement %d: %v", i, err)
		}
	}
	got, _ := m.GetRemainingShared(ctx, "pool1")
	if got != 0 {
		t.Fatalf("want 0 after 3 decrements from 3, got %d", got)
	}
}

// DecrementRemainingBorrow

func TestMemory_DecrementBorrow_NotFound(t *testing.T) {
	m := NewMemory()
	err := m.DecrementRemainingBorrow(ctx, "pool1", "profile1")
	if err == nil {
		t.Fatal("want error decrementing missing borrow pool")
	}
}

func TestMemory_DecrementBorrow_Decrements(t *testing.T) {
	m := NewMemory()
	m.SetBorrow("pool1", "profile1", 5)
	if err := m.DecrementRemainingBorrow(ctx, "pool1", "profile1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := m.GetRemainingBorrow(ctx, "pool1", "profile1")
	if got != 4 {
		t.Fatalf("want 4 after decrement, got %d", got)
	}
}

// Borrow decrements are per-profile — profile2's counter is unaffected
func TestMemory_DecrementBorrow_ProfileIsolation(t *testing.T) {
	m := NewMemory()
	m.SetBorrow("pool1", "profile1", 10)
	m.SetBorrow("pool1", "profile2", 10)

	m.DecrementRemainingBorrow(ctx, "pool1", "profile1")
	m.DecrementRemainingBorrow(ctx, "pool1", "profile1")

	p1, _ := m.GetRemainingBorrow(ctx, "pool1", "profile1")
	p2, _ := m.GetRemainingBorrow(ctx, "pool1", "profile2")

	if p1 != 8 {
		t.Fatalf("profile1: want 8, got %d", p1)
	}
	if p2 != 10 {
		t.Fatalf("profile2: want 10 (unaffected), got %d", p2)
	}
}

// ResetShared / ResetBorrow return errors (unimplemented)

func TestMemory_ResetShared_Unimplemented(t *testing.T) {
	m := NewMemory()
	err := m.ResetShared(ctx, "pool1", 100, 3600)
	if err == nil {
		t.Fatal("want error from unimplemented ResetShared")
	}
}

func TestMemory_ResetBorrow_Unimplemented(t *testing.T) {
	m := NewMemory()
	err := m.ResetBorrow(ctx, "pool1", "profile1", 100, 3600)
	if err == nil {
		t.Fatal("want error from unimplemented ResetBorrow")
	}
}
