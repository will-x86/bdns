package db

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
)

const migrationsDir = "../../migrations/"

func initTestDB(t *testing.T) *SQLiteStores {
	t.Helper()
	if err := InitDB(zerolog.Nop(), ":memory:", migrationsDir); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	return NewStores(GetDB())
}

func TestProfileExists_NotFound(t *testing.T) {
	s := initTestDB(t)

	exists, err := s.ProfileExists(context.Background(), "doesnotexist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for unknown profile, got true")
	}
}

func TestProfileExists_Found(t *testing.T) {
	s := initTestDB(t)

	userID, err := CreateUser("Europe/London")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	profileID, err := CreateProfile(userID, "test-laptop")
	if err != nil {
		t.Fatalf("CreateProfile: %v", err)
	}

	exists, err := s.ProfileExists(context.Background(), profileID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Errorf("expected true for profile %q, got false", profileID)
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
