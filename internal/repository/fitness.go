package repository

import (
	"database/sql"
	"time"
	
	_ "modernc.org/sqlite"
)

type FitnessRepo struct {
	DB *sql.DB
}

func NewFitnessRepo(db *sql.DB) *FitnessRepo {
	return &FitnessRepo{DB: db}
}

func InitTables(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS requests_history (
		    id integer PRIMARY KEY NOT NULL,
		    muscle TEXT NOT NULL,
		    requested_at TIMESTAMP NOT NULL
		)
`)
	if err != nil {
		//	logger
	}
}

func (r *FitnessRepo) SaveRequest(muscle string) error {
	_, err := r.DB.Exec(`
		INSERT INTO request_history (muscle, requested_at)
		VALUES ($1, $2)`, muscle, time.Now().Local())
	return err
}
