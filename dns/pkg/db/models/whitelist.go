package models

type PermanentWhitelist struct {
	ProfileID string `db:"profile_id" json:"profile_id"`
	Domain    string `db:"domain" json:"domain"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type TemporaryWhitelist struct {
	ProfileID string `db:"profile_id" json:"profile_id"`
	Domain    string `db:"domain" json:"domain"`
	ExpiresAt int64  `db:"expires_at" json:"expires_at"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type CategoryBlock struct {
	ProfileID string `db:"profile_id" json:"profile_id"`
	Category  string `db:"category" json:"category"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type UserFriend struct {
	UserID    string `db:"user_id" json:"user_id"`
	FriendID  string `db:"friend_id" json:"friend_id"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}
