package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/will-x86/bdns/dns/pkg/db/models"
)

// Impls stores from pkg/rule/context.go & CategoryResolver
type SQLiteStores struct {
	db *sqlx.DB
}

func NewStores(db *sqlx.DB) *SQLiteStores {
	return &SQLiteStores{db: db}
}

// Returns true if profile has *perm* whitelisted domain.
func (s *SQLiteStores) IsPermanentlyWhitelisted(ctx context.Context, profileID, domain string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM permanent_whitelists WHERE profile_id = ? AND domain = ? LIMIT 1`,
		profileID, domain,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// True if temp whitelisted
func (s *SQLiteStores) IsTemporarilyWhitelisted(ctx context.Context, profileID, domain string, now time.Time) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM temporary_whitelists WHERE profile_id = ? AND domain = ? AND expires_at > ? LIMIT 1`,
		profileID, domain, now.Unix(),
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// Hard block on entire category for a profile, e.g. "porn" / "gambling"
func (s *SQLiteStores) IsCategoryBlocked(ctx context.Context, profileID, category string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM user_category_blocks WHERE profile_id = ? AND category = ? LIMIT 1`,
		profileID, category,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *SQLiteStores) ProfileExists(ctx context.Context, profileID string) (bool, error) {
	var exists int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM profiles WHERE id = ? LIMIT 1`,
		profileID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *SQLiteStores) GetProfileWithUser(ctx context.Context, profileID string) (*models.Profile, *models.User, error) {
	var profile models.Profile
	var user models.User
	err := s.db.QueryRowContext(ctx,
		`SELECT p.id, p.user_id, p.name, p.created_at,
		        u.id, u.timezone, u.created_at
		 FROM profiles p
		 JOIN users u ON u.id = p.user_id
		 WHERE p.id = ? LIMIT 1`,
		profileID,
	).Scan(
		&profile.ID, &profile.UserID, &profile.Name, &profile.CreatedAt,
		&user.ID, &user.Timezone, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	return &profile, &user, nil
}

// Check DB if domain is categorized, if not ""
func (s *SQLiteStores) ResolveCategory(ctx context.Context, domain string) (string, error) {
	log.Printf("resolving category for domain: %s", domain)
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

func (s *SQLiteStores) GetTimeBlocks(ctx context.Context, profileID, category string) ([]models.TimeBlock, error) {
	log.Printf("Grabbing timeblocks for profile ID %s with category %s", profileID, category)
	var timeblocks []models.TimeBlock
	err := db.Select(&timeblocks, "SELECT * FROM user_time_blocks WHERE profile_id=? AND category=?", profileID, category)
	fmt.Printf("%T", err)
	if err != nil {
		return []models.TimeBlock{}, err
	}

	return timeblocks, nil
}
