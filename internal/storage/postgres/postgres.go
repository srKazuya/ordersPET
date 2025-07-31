package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"
)

type Storage struct {
	db *sql.DB
}

func New(cfg Config) (*Storage, error) {
	const op = "storage.postgres.NewStrorage"

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("%s migrate:%w", op, err)
	}

	return &Storage{db: db}, nil
}
