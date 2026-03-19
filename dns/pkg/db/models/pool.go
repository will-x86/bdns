package models

type FriendPool struct {
	ID         string `db:"id"`
	CreatedBy  string `db:"created_by"`
	Name       string `db:"name"`
	PoolMode   string `db:"pool_mode"`
	TotalLimit int64  `db:"total_limit"`
	CreatedAt  int64  `db:"created_at"`
}
