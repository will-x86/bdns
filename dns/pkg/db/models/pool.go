package models

type FriendPool struct {
	ID         string `db:"id" json:"id"`
	CreatedBy  string `db:"created_by" json:"created_by"`
	Name       string `db:"name" json:"name"`
	PoolMode   string `db:"pool_mode" json:"pool_mode"`
	TotalLimit int64  `db:"total_limit" json:"total_limit"`
	CreatedAt  int64  `db:"created_at" json:"created_at"`
}

type FriendPoolMembers struct {
	PoolID    string `db:"pool_id" json:"pool_id"`
	ProfileID string `db:"profile_id" json:"profile_id"`
	JoinedAt  int64  `db:"joined_at" json:"joined_at"`
}
type FriendPoolCategoryBlocks struct {
	PoolID    string `db:"pool_id" json:"pool_id"`
	Category  string `db:"category" json:"category"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}
