package models

type User struct {
	ID        string `db:"id"`
	Timezone  string `db:"timezone"` // 'Europe/London',
	CreatedAt int64  `db:"created_at"`
}
type Profile struct {
	ID        string `db:"id"`
	UserID    string `db:"user_id"`
	Name      string `db:"name"`
	CreatedAt int64  `db:"created_at"`
}
