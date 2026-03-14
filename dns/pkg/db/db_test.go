package db

import (
	"context"
	"testing"
)

const migrationsDir = "../../migrations/"

func initTestDB(t *testing.T) *SQLiteStores {
	t.Helper()
	if err := InitDB(":memory:", migrationsDir); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return NewStores(GetDB())
}

func TestUserExists_NotFound(t *testing.T) {
	s := initTestDB(t)

	exists, err := s.UserExists(context.Background(), "doesnotexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for unknown user, got true")
	}
}

func TestUserExists_Found(t *testing.T) {
	s := initTestDB(t)

	id, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	exists, err := s.UserExists(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true for user %q, got false", id)
	}
}

func TestCreateUser_ReturnsID(t *testing.T) {
	_ = initTestDB(t)

	id, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id == "" {
		t.Error("expected non-empty ID from CreateUser")
	}
}

func TestCreateUser_UniqueIDs(t *testing.T) {
	_ = initTestDB(t)

	a, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser (1): %v", err)
	}
	b, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser (2): %v", err)
	}
	if a == b {
		t.Errorf("expected unique IDs, both returned %q", a)
	}
}
