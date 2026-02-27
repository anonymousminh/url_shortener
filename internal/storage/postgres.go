package storage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrURLNotFound = errors.New("URL not found")

type PostgresStore struct {
	db *pgxpool.Pool
}

func NewPostgresStorage(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) SaveURL(ctx context.Context, originalURL string) (int64, error) {
	query := "INSERT INTO urls (original_url) VALUES ($1) RETURNING id"
	var id int64
	err := s.db.QueryRow(ctx, query, originalURL).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *PostgresStore) UpdateShortCode(ctx context.Context, id int64, shortCode string) error {
	query := "UPDATE urls SET short_code = $1 WHERE id = $2"
	ct, err := s.db.Exec(ctx, query, shortCode, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("no rows affected")
	}
	return nil
}

func (s *PostgresStore) GetURLByShortCode(ctx context.Context, shortCode string) (string, error) {
	query := "SELECT original_url FROM urls WHERE short_code = $1"
	var originalURL string
	err := s.db.QueryRow(ctx, query, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", err
	}
	return originalURL, nil
}
