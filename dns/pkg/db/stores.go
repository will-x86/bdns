package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// Impls stores from pkg/rule/context.go & CategoryResolver
type SQLiteStores struct {
	db *sqlx.DB
}

func NewStores(db *sqlx.DB) *SQLiteStores {
	return &SQLiteStores{db: db}
}

// Returns true if user has *perm* whitelisted domain.
func (s *SQLiteStores) IsPermanentlyWhitelisted(ctx context.Context, userID, domain string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM permanent_whitelists WHERE user_id = ? AND domain = ? LIMIT 1`,
		userID, domain,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// True if temp whitelisted
func (s *SQLiteStores) IsTemporarilyWhitelisted(ctx context.Context, userID, domain string, now time.Time) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM temporary_whitelists WHERE user_id = ? AND domain = ? AND expires_at > ? LIMIT 1`,
		userID, domain, now.Unix(),
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// Hard block on entire profile, e.g. "porn" / "gambling"
func (s *SQLiteStores) IsCategoryBlocked(ctx context.Context, userID, category string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM user_category_blocks WHERE user_id = ? AND category = ? LIMIT 1`,
		userID, category,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// Check DB if domain is categorized, if not ""
func (s *SQLiteStores) ResolveCategory(ctx context.Context, domain string) (string, error) {
	var category string
	err := s.db.QueryRowContext(ctx,
		`SELECT category FROM blocklist_entries WHERE domain = ? LIMIT 1`,
		domain,
	).Scan(&category)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return category, err
}
