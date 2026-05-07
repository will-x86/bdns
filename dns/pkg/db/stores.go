package db

import (
	"context"
	"database/sql"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
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

func (s *SQLiteStores) ListProfiles(ctx context.Context, userID string) ([]models.Profile, error) {
	var profiles []models.Profile
	err := s.db.SelectContext(ctx, &profiles,
		`SELECT id, user_id, name, created_at FROM profiles WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s *SQLiteStores) GetProfile(ctx context.Context, profileID string) (*models.Profile, error) {
	var profile models.Profile
	err := s.db.GetContext(ctx, &profile,
		`SELECT id, user_id, name, created_at FROM profiles WHERE id = ?`,
		profileID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *SQLiteStores) UpdateProfile(ctx context.Context, profileID, name string) (*models.Profile, error) {
	var profile models.Profile
	err := s.db.GetContext(ctx, &profile,
		`UPDATE profiles SET name = ? WHERE id = ? RETURNING id, user_id, name, created_at`,
		name, profileID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *SQLiteStores) DeleteProfile(ctx context.Context, profileID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, profileID)
	return err
}

func (s *SQLiteStores) ListPermanentWhitelists(ctx context.Context, profileID string) ([]string, error) {
	var domains []string
	err := s.db.SelectContext(ctx, &domains,
		`SELECT domain FROM permanent_whitelists WHERE profile_id = ?`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	return domains, nil
}

func (s *SQLiteStores) AddPermanentWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO permanent_whitelists (profile_id, domain) VALUES (?, ?)`,
		profileID, domain,
	)
	return err
}

func (s *SQLiteStores) RemovePermanentWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM permanent_whitelists WHERE profile_id = ? AND domain = ?`,
		profileID, domain,
	)
	return err
}

type TempWhitelist struct {
	Domain    string `db:"domain"`
	ExpiresAt int    `db:"expires_at"`
}

func (s *SQLiteStores) ListTemporaryWhitelists(ctx context.Context, profileID string) ([]TempWhitelist, error) {
	var whitelists []TempWhitelist
	err := s.db.SelectContext(ctx, &whitelists,
		`SELECT domain, expires_at FROM temporary_whitelists WHERE profile_id = ?`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	return whitelists, nil
}

func (s *SQLiteStores) AddTemporaryWhitelist(ctx context.Context, profileID, domain string, expiresAt int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO temporary_whitelists (profile_id, domain, expires_at) VALUES (?, ?, ?)`,
		profileID, domain, expiresAt,
	)
	return err
}

func (s *SQLiteStores) RemoveTemporaryWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM temporary_whitelists WHERE profile_id = ? AND domain = ?`,
		profileID, domain,
	)
	return err
}

func (s *SQLiteStores) ListBlockedCategories(ctx context.Context, profileID string) ([]string, error) {
	var categories []string
	err := s.db.SelectContext(ctx, &categories,
		`SELECT category FROM user_category_blocks WHERE profile_id = ?`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *SQLiteStores) BlockCategory(ctx context.Context, profileID, category string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_category_blocks (profile_id, category) VALUES (?, ?)`,
		profileID, category,
	)
	return err
}

func (s *SQLiteStores) UnblockCategory(ctx context.Context, profileID, category string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM user_category_blocks WHERE profile_id = ? AND category = ?`,
		profileID, category,
	)
	return err
}

func (s *SQLiteStores) ListTimeBlocks(ctx context.Context, profileID string) ([]models.TimeBlock, error) {
	var timeblocks []models.TimeBlock
	err := s.db.SelectContext(ctx, &timeblocks,
		`SELECT profile_id, category, start_time, end_time, day, created_at FROM user_time_blocks WHERE profile_id = ?`,
		profileID,
	)
	if err != nil {
		return nil, err
	}
	return timeblocks, nil
}

func (s *SQLiteStores) CreateTimeBlock(ctx context.Context, profileID, category string, startTime, endTime, day int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO user_time_blocks (profile_id, category, start_time, end_time, day) VALUES (?, ?, ?, ?, ?)`,
		profileID, category, startTime, endTime, day,
	)
	return err
}

func (s *SQLiteStores) DeleteTimeBlock(ctx context.Context, profileID, category string, startTime, endTime, day int) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM user_time_blocks WHERE profile_id = ? AND category = ? AND start_time = ? AND end_time = ? AND day = ?`,
		profileID, category, startTime, endTime, day,
	)
	return err
}

func (s *SQLiteStores) DeleteTimeBlockByProfile(ctx context.Context, profileID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_time_blocks WHERE profile_id = ?`, profileID)
	return err
}

func (s *SQLiteStores) ListPoolsForUser(ctx context.Context, userID string) ([]models.FriendPool, error) {
	var pools []models.FriendPool
	err := s.db.SelectContext(ctx, &pools,
		`SELECT id, created_by, pool_mode, total_limit FROM friend_pools WHERE created_by = ?`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return pools, nil
}

func (s *SQLiteStores) CreatePool(ctx context.Context, userID, name, poolMode string, limit int64) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO friend_pools (created_by, name, pool_mode, total_limit) VALUES (?, ?, ?, ?) RETURNING id`,
		userID, name, poolMode, limit,
	).Scan(&id)
	return id, err
}

func (s *SQLiteStores) DeletePool(ctx context.Context, poolID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM friend_pools WHERE id = ?`, poolID)
	return err
}

func (s *SQLiteStores) JoinPool(ctx context.Context, poolID, profileID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO friend_pool_members (pool_id, profile_id) VALUES (?, ?)`,
		poolID, profileID,
	)
	return err
}

func (s *SQLiteStores) LeavePool(ctx context.Context, poolID, profileID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM friend_pool_members WHERE pool_id = ? AND profile_id = ?`,
		poolID, profileID,
	)
	return err
}

func (s *SQLiteStores) ListPoolMembers(ctx context.Context, poolID string) ([]models.FriendPoolMembers, error) {
	var members []models.FriendPoolMembers
	err := s.db.SelectContext(ctx, &members,
		`SELECT pool_id, profile_id FROM friend_pool_members WHERE pool_id = ?`,
		poolID,
	)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (s *SQLiteStores) ListPoolCategoryBlocks(ctx context.Context, poolID string) ([]string, error) {
	var categories []string
	err := s.db.SelectContext(ctx, &categories,
		`SELECT category FROM friend_pool_category_blocks WHERE pool_id = ?`,
		poolID,
	)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (s *SQLiteStores) AddPoolCategoryBlock(ctx context.Context, poolID, category string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO friend_pool_category_blocks (pool_id, category) VALUES (?, ?)`,
		poolID, category,
	)
	return err
}

func (s *SQLiteStores) RemovePoolCategoryBlock(ctx context.Context, poolID, category string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM friend_pool_category_blocks WHERE pool_id = ? AND category = ?`,
		poolID, category,
	)
	return err
}

func (s *SQLiteStores) GetPoolCredits(ctx context.Context, poolID string) (int64, error) {
	var limit int64
	err := s.db.QueryRowContext(ctx,
		`SELECT total_limit FROM friend_pools WHERE id = ?`,
		poolID,
	).Scan(&limit)
	return limit, err
}

func (s *SQLiteStores) ListCategories(ctx context.Context) ([]string, error) {
	var categories []string
	err := s.db.SelectContext(ctx, &categories,
		`SELECT DISTINCT category FROM blocklist_sources`,
	)
	if err != nil {
		return nil, err
	}
	return categories, nil
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

func (s *SQLiteStores) CreateUser(ctx context.Context, timezone string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, "INSERT INTO users (timezone) VALUES (?) RETURNING id", timezone).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}
func (s *SQLiteStores) UserExists(ctx context.Context, userID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SQLiteStores) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.GetContext(ctx, &user, "SELECT id, timezone, created_at FROM users WHERE id = ?", userID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *SQLiteStores) UpdateUser(ctx context.Context, userID, timezone string) (*models.User, error) {
	_, err := s.db.ExecContext(ctx, "UPDATE users SET timezone = ? WHERE id = ?", timezone, userID)
	if err != nil {
		return nil, err
	}
	return s.GetUser(ctx, userID)
}
