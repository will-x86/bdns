package models

type User struct {
	ID        string `db:"id" json:"id"`
	Timezone  string `db:"timezone" json:"timezone"` // 'Europe/London',
	APIToken  string `db:"api_token" json:"-"`       // secret, never serialised in list/get responses
	CreatedAt int64  `db:"created_at" json:"created_at"`
}
type Profile struct {
	ID        string `db:"id" json:"id"`
	UserID    string `db:"user_id" json:"user_id"`
	Name      string `db:"name" json:"name"`
	CreatedAt int64  `db:"created_at" json:"created_at"`
}

type AllMembers struct {
	PoolID     string
	CreatedBy  string
	ProfileID  string
	PoolMode   string
	TotalLimit int64
	Timezone   string
	UserID     string
}
