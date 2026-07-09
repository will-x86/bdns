package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/will-x86/bdns/dns/pkg/db/models"
)

type Repo struct {
	db *sqlx.DB
}

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateUser(ctx context.Context, timezone, token string) (models.User, error) {
	var u models.User
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO users (timezone, api_token) VALUES (?, ?)
		 RETURNING id, timezone, api_token, created_at`,
		timezone, token,
	).StructScan(&u)
	return u, err
}

func (r *Repo) UserByToken(ctx context.Context, token string) (*models.User, error) {
	var u models.User
	err := r.db.GetContext(ctx, &u,
		`SELECT id, timezone, api_token, created_at FROM users WHERE api_token = ?`, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) UserExists(ctx context.Context, userID string) (bool, error) {
	return r.exists(ctx, `SELECT 1 FROM users WHERE id = ? LIMIT 1`, userID)
}

func (r *Repo) UpdateUserTimezone(ctx context.Context, userID, timezone string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET timezone = ? WHERE id = ?`, timezone, userID)
	return err
}

func (r *Repo) SetUserToken(ctx context.Context, userID, token string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET api_token = ? WHERE id = ?`, token, userID)
	return err
}

func (r *Repo) DeleteUser(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

func (r *Repo) ListProfiles(ctx context.Context, userID string) ([]models.Profile, error) {
	profiles := []models.Profile{}
	err := r.db.SelectContext(ctx, &profiles,
		`SELECT id, user_id, name, created_at FROM profiles WHERE user_id = ? ORDER BY created_at`, userID)
	return profiles, err
}

func (r *Repo) CreateProfile(ctx context.Context, userID, name string) (models.Profile, error) {
	var p models.Profile
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO profiles (user_id, name) VALUES (?, ?)
		 RETURNING id, user_id, name, created_at`,
		userID, name,
	).StructScan(&p)
	return p, err
}

func (r *Repo) GetProfile(ctx context.Context, profileID string) (*models.Profile, error) {
	var p models.Profile
	err := r.db.GetContext(ctx, &p,
		`SELECT id, user_id, name, created_at FROM profiles WHERE id = ?`, profileID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) UpdateProfileName(ctx context.Context, profileID, name string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE profiles SET name = ? WHERE id = ?`, name, profileID)
	return err
}

func (r *Repo) DeleteProfile(ctx context.Context, profileID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM profiles WHERE id = ?`, profileID)
	return err
}

func (r *Repo) ListPermanentWhitelist(ctx context.Context, profileID string) ([]models.PermanentWhitelist, error) {
	out := []models.PermanentWhitelist{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT profile_id, domain, created_at FROM permanent_whitelists WHERE profile_id = ? ORDER BY domain`, profileID)
	return out, err
}

func (r *Repo) AddPermanentWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO permanent_whitelists (profile_id, domain) VALUES (?, ?)`, profileID, domain)
	return err
}

func (r *Repo) DeletePermanentWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM permanent_whitelists WHERE profile_id = ? AND domain = ?`, profileID, domain)
	return err
}

func (r *Repo) ListTemporaryWhitelist(ctx context.Context, profileID string) ([]models.TemporaryWhitelist, error) {
	out := []models.TemporaryWhitelist{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT profile_id, domain, expires_at, created_at FROM temporary_whitelists WHERE profile_id = ? ORDER BY expires_at`, profileID)
	return out, err
}

func (r *Repo) AddTemporaryWhitelist(ctx context.Context, profileID, domain string, expiresAt int64) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO temporary_whitelists (profile_id, domain, expires_at) VALUES (?, ?, ?)
		 ON CONFLICT(profile_id, domain) DO UPDATE SET expires_at = excluded.expires_at`,
		profileID, domain, expiresAt)
	return err
}

func (r *Repo) DeleteTemporaryWhitelist(ctx context.Context, profileID, domain string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM temporary_whitelists WHERE profile_id = ? AND domain = ?`, profileID, domain)
	return err
}

func (r *Repo) ListCategoryBlocks(ctx context.Context, profileID string) ([]models.CategoryBlock, error) {
	out := []models.CategoryBlock{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT profile_id, category, created_at FROM user_category_blocks WHERE profile_id = ? ORDER BY category`, profileID)
	return out, err
}

func (r *Repo) AddCategoryBlock(ctx context.Context, profileID, category string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_category_blocks (profile_id, category) VALUES (?, ?)`, profileID, category)
	return err
}

func (r *Repo) DeleteCategoryBlock(ctx context.Context, profileID, category string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_category_blocks WHERE profile_id = ? AND category = ?`, profileID, category)
	return err
}

func (r *Repo) ListTimeBlocks(ctx context.Context, profileID string) ([]models.TimeBlock, error) {
	out := []models.TimeBlock{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT profile_id, category, start_time, end_time, day, created_at
		 FROM user_time_blocks WHERE profile_id = ? ORDER BY day, start_time`, profileID)
	return out, err
}

func (r *Repo) AddTimeBlock(ctx context.Context, tb models.TimeBlock) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_time_blocks (profile_id, category, start_time, end_time, day)
		 VALUES (?, ?, ?, ?, ?)`,
		tb.ProfileID, tb.Category, tb.StartTime, tb.EndTime, tb.Day)
	return err
}

func (r *Repo) DeleteTimeBlock(ctx context.Context, profileID, category string, day, start, end int) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_time_blocks
		 WHERE profile_id = ? AND category = ? AND day = ? AND start_time = ? AND end_time = ?`,
		profileID, category, day, start, end)
	return err
}

func (r *Repo) ListCategories(ctx context.Context) ([]string, error) {
	out := []string{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT DISTINCT category FROM blocklist_sources ORDER BY category`)
	return out, err
}

func (r *Repo) ListFriends(ctx context.Context, userID string) ([]models.User, error) {
	out := []models.User{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT u.id, u.timezone, u.api_token, u.created_at
		 FROM user_friends f JOIN users u ON u.id = f.friend_id
		 WHERE f.user_id = ? ORDER BY u.created_at`, userID)
	return out, err
}

func (r *Repo) IsFriend(ctx context.Context, userID, friendID string) (bool, error) {
	return r.exists(ctx,
		`SELECT 1 FROM user_friends WHERE user_id = ? AND friend_id = ? LIMIT 1`, userID, friendID)
}

func (r *Repo) AddFriend(ctx context.Context, userID, friendID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_friends (user_id, friend_id) VALUES (?, ?)`, userID, friendID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_friends (user_id, friend_id) VALUES (?, ?)`, friendID, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repo) DeleteFriend(ctx context.Context, userID, friendID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM user_friends WHERE (user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)`,
		userID, friendID, friendID, userID)
	return err
}

func (r *Repo) ListPoolsForUser(ctx context.Context, userID string) ([]models.FriendPool, error) {
	out := []models.FriendPool{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT DISTINCT fp.id, fp.created_by, fp.name, fp.pool_mode, fp.total_limit, fp.created_at
		 FROM friend_pools fp
		 LEFT JOIN friend_pool_members fpm ON fpm.pool_id = fp.id
		 LEFT JOIN profiles p ON p.id = fpm.profile_id
		 WHERE fp.created_by = ? OR p.user_id = ?
		 ORDER BY fp.created_at`, userID, userID)
	return out, err
}

func (r *Repo) CreatePool(ctx context.Context, createdBy, name, mode string, totalLimit int64) (models.FriendPool, error) {
	var p models.FriendPool
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO friend_pools (created_by, name, pool_mode, total_limit) VALUES (?, ?, ?, ?)
		 RETURNING id, created_by, name, pool_mode, total_limit, created_at`,
		createdBy, name, mode, totalLimit,
	).StructScan(&p)
	return p, err
}

func (r *Repo) GetPool(ctx context.Context, poolID string) (*models.FriendPool, error) {
	var p models.FriendPool
	err := r.db.GetContext(ctx, &p,
		`SELECT id, created_by, name, pool_mode, total_limit, created_at FROM friend_pools WHERE id = ?`, poolID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repo) UpdatePool(ctx context.Context, poolID, name string, totalLimit int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE friend_pools SET name = ?, total_limit = ? WHERE id = ?`, name, totalLimit, poolID)
	return err
}

func (r *Repo) DeletePool(ctx context.Context, poolID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM friend_pools WHERE id = ?`, poolID)
	return err
}

func (r *Repo) ListPoolMembers(ctx context.Context, poolID string) ([]models.FriendPoolMembers, error) {
	out := []models.FriendPoolMembers{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT pool_id, profile_id, joined_at FROM friend_pool_members WHERE pool_id = ? ORDER BY joined_at`, poolID)
	return out, err
}

func (r *Repo) AddPoolMember(ctx context.Context, poolID, profileID string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO friend_pool_members (pool_id, profile_id) VALUES (?, ?)`, poolID, profileID)
	return err
}

func (r *Repo) DeletePoolMember(ctx context.Context, poolID, profileID string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM friend_pool_members WHERE pool_id = ? AND profile_id = ?`, poolID, profileID)
	return err
}

func (r *Repo) UserInPool(ctx context.Context, poolID, userID string) (bool, error) {
	return r.exists(ctx,
		`SELECT 1 FROM friend_pool_members fpm
		 JOIN profiles p ON p.id = fpm.profile_id
		 WHERE fpm.pool_id = ? AND p.user_id = ? LIMIT 1`, poolID, userID)
}

func (r *Repo) ListPoolCategoryBlocks(ctx context.Context, poolID string) ([]models.FriendPoolCategoryBlocks, error) {
	out := []models.FriendPoolCategoryBlocks{}
	err := r.db.SelectContext(ctx, &out,
		`SELECT pool_id, category, created_at FROM friend_pool_category_blocks WHERE pool_id = ? ORDER BY category`, poolID)
	return out, err
}

func (r *Repo) AddPoolCategoryBlock(ctx context.Context, poolID, category string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO friend_pool_category_blocks (pool_id, category) VALUES (?, ?)`, poolID, category)
	return err
}

func (r *Repo) DeletePoolCategoryBlock(ctx context.Context, poolID, category string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM friend_pool_category_blocks WHERE pool_id = ? AND category = ?`, poolID, category)
	return err
}

func (r *Repo) exists(ctx context.Context, query string, args ...any) (bool, error) {
	var one int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&one)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
