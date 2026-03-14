package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func runMigrations(db *sql.DB, migrationsPath string) {
	if err := goose.SetDialect("sqlite"); err != nil {
		log.Fatalf("error setting database dialiect to sqlite : %v", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		log.Fatalf("error migrating database : %v", err)
	}

}

func InitDB(dbPath, migrationsDir string) error {
	var err error
	db, err = sqlx.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	runMigrations(db.DB, migrationsDir)
	return nil
}

func Seed() {
	path := "./seed.sql"

	c, e := os.ReadFile(path)
	if e != nil {
		log.Fatalf("error reading seed file, path: %s, err - %v", path, e)
	}
	sql := string(c)
	_, err := db.Exec(sql)
	if err != nil {
		log.Fatalf("error executing seed file, path: %s, err - %v", path, err)
	}
}
func GetDB() *sqlx.DB {
	return db
}

func CreateUser(timezone string) (string, error) {
	var id string
	err := db.QueryRow(
		`INSERT INTO users (timezone) VALUES (?) RETURNING id`,
		timezone,
	).Scan(&id)
	return id, err
}
