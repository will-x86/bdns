package db

import (
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func InitDB(dbPath string) error {
	var err error
	db, err = sqlx.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

type User struct {
	ID        string `db:"id"`
	CreatedAt int64  `db:"created_at"`
}

func CreateUser() (string, error) {
	var id string
	err := db.QueryRow(
		`INSERT INTO users DEFAULT VALUES RETURNING id`,
	).Scan(&id)
	return id, err
}
