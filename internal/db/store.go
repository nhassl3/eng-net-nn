package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: New(pool),
		pool:    pool,
	}
}

func (s *Store) ExecTx(ctx context.Context, fn func(*Queries) error) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquiring connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := fn(s.WithTx(tx)); err != nil {
		errTx := fmt.Errorf("executing transaction: %w", err)
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rolling back transaction: %w", errors.Join(errTx, rbErr))
		}
		return errTx
	}

	return tx.Commit(ctx)
}

func (s *Store) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *Store) BeginTx(ctx context.Context) (*Queries, pgx.Tx, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("beginning transaction: %w", err)
	}
	return s.WithTx(tx), tx, nil
}
