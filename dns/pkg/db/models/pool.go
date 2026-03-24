package models

type FriendPool struct {
	ID         string `db:"id"`
	CreatedBy  string `db:"created_by"`
	Name       string `db:"name"`
	PoolMode   string `db:"pool_mode"`
	TotalLimit int64  `db:"total_limit"`
	CreatedAt  int64  `db:"created_at"`
}

type FriendPoolMembers struct {
	PoolID    string `db:"pool_id"`
	ProfileID string `db:"profile_id"`
	JoinedAt  int64  `db:"joined_at"`
}
type FriendPoolCategoryBlocks struct {
	PoolID    string `db:"pool_id"`
	Category  string `db:"category"`
	CreatedAt int64  `db:"created_at"`
}
