package db

import (
	"database/sql"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/will-x86/bdns/dns/pkg/db/models"
	_ "modernc.org/sqlite"
)

var db *sqlx.DB

func runMigrations(log zerolog.Logger, db *sql.DB, migrationsPath string) {
	if err := goose.SetDialect("sqlite"); err != nil {
		log.Fatal().Err(err).Msg("error setting database dialect to sqlite")
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		log.Fatal().Err(err).Str("path", migrationsPath).Msg("error migrating database")
	}

}

func InitDB(log zerolog.Logger, dbPath, migrationsDir string) error {
	log = log.With().Str("component", "initdb").Logger()
	var err error
	db, err = sqlx.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	runMigrations(log, db.DB, migrationsDir)
	return nil
}

func Seed(log zerolog.Logger) {
	log = log.With().Str("component", "db-seed").Logger()
	path := "./seed.sql"

	c, e := os.ReadFile(path)
	if e != nil {
		log.Fatal().Err(e).Str("path", path).Msg("error reading seed file")
	}
	sql := string(c)
	_, err := db.Exec(sql)
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("error executing seed file")
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

func CreateProfile(userID, name string) (string, error) {
	var id string
	err := db.QueryRow(
		`INSERT INTO profiles (user_id, name) VALUES (?, ?) RETURNING id`,
		userID, name,
	).Scan(&id)
	return id, err
}
func GetAllFriendPoolMembers() ([]models.FriendPoolMembers, error) {
	var poolMembers []models.FriendPoolMembers
	err := db.Select(&poolMembers, "SELECT * FROM friend_pool_members")
	if err != nil {
		return []models.FriendPoolMembers{}, err
	}
	return poolMembers, nil

}
func GetAllFriendPools() ([]models.FriendPool, error) {
	var friendPools []models.FriendPool
	err := db.Select(&friendPools, "SELECT * FROM friend_pools")
	if err != nil {
		return []models.FriendPool{}, err
	}
	return friendPools, nil
}
