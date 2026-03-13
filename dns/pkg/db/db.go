package db

import (
	"database/sql"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func runMigrations(db *sql.DB, migrationsPath string) {
	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		panic(err)
	}

}

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
	runMigrations(db.DB, "./migrations/")
	return nil
}

func Seed() {
	path := "./seed.sql"

	c, e := os.ReadFile(path)
	if e != nil {
		panic(e)
	}
	sql := string(c)
	_, err := db.Exec(sql)
	if err != nil {
		panic(err)
	}
}
func GetDB() *sqlx.DB {
	return db
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
func UserExists(userID string) (bool, error) {
	var exists int
	err := db.QueryRow(
		`SELECT 1 FROM users WHERE id = ? LIMIT 1`,
		userID,
	).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
