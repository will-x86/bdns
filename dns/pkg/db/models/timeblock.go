package models

type TimeBlock struct {
	ProfileID string `db:"profile_id"`
	Category  string `db:"category"`
	StartTime int    `db:"start_time"`
	EndTime   int    `db:"end_time"`
	Day       int    `db:"day"`
	CreatedAt int    `db:"created_at"`
}
