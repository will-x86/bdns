package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
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
	log := zerolog.Ctx(ctx).With().Str("component", "db-stores").Logger()
	log.Debug().Str("domain", domain).Msg("resolving category")
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
	log := zerolog.Ctx(ctx).With().Str("component", "db-stores").Logger()
	log.Debug().Str("profileID", profileID).Str("category", category).Msg("getting timeblocks")
	var timeblocks []models.TimeBlock
	err := s.db.SelectContext(ctx, &timeblocks, "SELECT * FROM user_time_blocks WHERE profile_id=? AND category=?", profileID, category)
	if err != nil {
		return []models.TimeBlock{}, err
	}

	return timeblocks, nil
}
func (s *SQLiteStores) GetPool(ctx context.Context, poolID string) (models.FriendPool, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "db-stores").Logger()
	log.Debug().Str("poolID", poolID).Msg("getting pool")
	var pool models.FriendPool
	err := s.db.GetContext(ctx, &pool, "SELECT * FROM friend_pools WHERE id=?", poolID)
	if err != nil {
		return models.FriendPool{}, err
	}
	return pool, nil
}

// Checks if a cateogory is blocked for a given pool
func (s *SQLiteStores) PoolCategoryBlocked(ctx context.Context, poolID, category string) bool {
	log := zerolog.Ctx(ctx).With().Str("component", "db-stores-pool-cateogry-blocked").Logger()
	log.Debug().Str("poolID", poolID).Str("category", category).Msg("checking pool category block")
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM friend_pool_category_blocks WHERE pool_id=? AND category=?", poolID, category).Scan(&count)
	if err != nil {
		log.Error().Err(err).Msg("grabbing pool category blocked failed")
		return false
	}
	log.Trace().Str("poolID", poolID).Int("count of friend_pool_category_blocks", count).Send()
	return count > 0
}

/*
Add GetAllPoolMembersWithTimezones() to SQLiteStores.
Join with friend_pool_members, profiles, users, friend_pools to get:
pool_id, pool_mode, total_limit , profile_id & timezone
*/

func (s *SQLiteStores) GetAllPoolMembersWithTimezones(ctx context.Context) ([]models.AllMembers, error) {
	// want pool_id, pool_mode, total_limit, profile_id & timezone

	log := zerolog.Ctx(ctx).With().Str("component", "db-stores-GetAllPoolMembersWithTimezones").Logger()
	log.Debug().Time("now", time.Now()).Msg("Getting all pool members with timezones at")
	rows, err := s.db.Query(`
    SELECT 
        fpm.pool_id,
        fpm.profile_id,
        fp.pool_mode,
        fp.total_limit,
		fp.created_by,
        u.timezone,
		p.user_id
    FROM friend_pools AS fp
    INNER JOIN friend_pool_members AS fpm ON fp.id = fpm.pool_id
    INNER JOIN profiles AS p ON fpm.profile_id = p.id
    INNER JOIN users AS u ON p.user_id = u.id
`)
	if err != nil {
		return []models.AllMembers{}, err
	}

	var allMembers []models.AllMembers
	// iterate over each row
	for rows.Next() {
		var member models.AllMembers
		err = rows.Scan(
			&member.PoolID,
			&member.ProfileID,
			&member.PoolMode,
			&member.TotalLimit,
			&member.CreatedBy,
			&member.Timezone,
			&member.UserID,
		)
		if err != nil {
			return []models.AllMembers{}, err
		}
		if member.Timezone == "" {
			log.Fatal().Str("userid", member.UserID).Msg("no timezone for member")
		}
		log.Trace().Any("member gotten", member).Send()
		allMembers = append(allMembers, member)
	}
	// check the error from rows
	err = rows.Err()
	return allMembers, err
}
