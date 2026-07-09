package models

type TimeBlock struct {
	ProfileID string `db:"profile_id" json:"profile_id"`
	Category  string `db:"category" json:"category"`
	StartTime int    `db:"start_time" json:"start_time"`
	EndTime   int    `db:"end_time" json:"end_time"`
	Day       int    `db:"day" json:"day"`
	CreatedAt int    `db:"created_at" json:"created_at"`
}
