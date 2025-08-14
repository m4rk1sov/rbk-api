package repository

import (
	"database/sql"
	"errors"
	"time"
	
	_ "modernc.org/sqlite"
)

type FitnessRepo struct {
	DB *sql.DB
}

func NewFitnessRepo(db *sql.DB) *FitnessRepo {
	return &FitnessRepo{DB: db}
}

func (r *FitnessRepo) InitTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS requests_history (
		    id INTEGER PRIMARY KEY,
		    muscle TEXT NOT NULL,
		    requested_at TIMESTAMP NOT NULL
		);

		CREATE TABLE IF NOT EXISTS muscle_cache (
		    muscle TEXT PRIMARY KEY,
		    data TEXT NOT NULL,
		    fetched_at TIMESTAMP NOT NULL
		);
`)
	return err
}

func (r *FitnessRepo) SaveRequest(muscle string) error {
	_, err := r.DB.Exec(`
		INSERT INTO requests_history (muscle, requested_at)
		VALUES (?, ?)`, muscle, time.Now().Local())
	return err
}

func (r *FitnessRepo) GetCachedMuscle(muscle string, ttl time.Duration) (string, bool, error) {
	var data string
	var fetchedAt time.Time
	
	err := r.DB.QueryRow(
		`SELECT data, fetched_at FROM muscle_cache WHERE muscle = ?`,
		muscle,
	).Scan(&data, &fetchedAt)
	
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	
	if err != nil {
		return "", false, err
	}
	
	if time.Since(fetchedAt) <= ttl {
		return data, true, nil
	}
	
	return "", false, nil
}

func (r *FitnessRepo) SaveMuscleCache(muscle, data string) error {
	_, err := r.DB.Exec(`
		INSERT INTO muscle_cache (muscle, data, fetched_at)
		VALUES (?, ?, ?)
		ON CONFLICT(muscle) DO UPDATE SET
		           data = excluded.data,
		           fetched_at = excluded.fetched_at
`, muscle, data, time.Now())
	return err
}
