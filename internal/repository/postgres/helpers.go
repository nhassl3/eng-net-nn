package postgres

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

// mapNotFound reports whether err is pgx.ErrNoRows and, if so, returns
// notFoundErr in its place so it survives as a *domain.DomainError up
// through the service layer to handleError, instead of falling through to a
// generic wrapped error that maps to 500. ok is false when err is left
// untouched.
func mapNotFound(err error, notFoundErr error) (mapped error, ok bool) {
	if errors.Is(err, pgx.ErrNoRows) {
		return notFoundErr, true
	}
	return err, false
}

// mapConstraintErr centralizes Postgres constraint-violation codes into
// domain errors: unique-violation (23505) becomes alreadyExistsErr (pass nil
// to leave it unmapped), not-null and check violations (23502, 23514) become
// domain.ErrInvalidInput (400, instead of a raw 500). ok is false — err is
// returned untouched — for any other error, including a pgconn.PgError with
// an unhandled code.
func mapConstraintErr(err error, alreadyExistsErr error) (mapped error, ok bool) {
	pgErr, isPgErr := errors.AsType[*pgconn.PgError](err)
	if !isPgErr {
		return err, false
	}
	switch pgErr.Code {
	case "23505":
		if alreadyExistsErr != nil {
			return alreadyExistsErr, true
		}
	case "23502", "23514":
		return domain.ErrInvalidInput, true
	case "P0001":
		return domain.ErrUserAlreadyHasRole, true
	}
	return err, false
}

// pgTimeTZ extracts time.Time from a pgtype.Timestamptz value.
// Falls back to a zero value if the timestamp is not valid.
func pgTimeTZ(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}

// uuid2String converts uuid to string type
func uuid2String(id uuid.UUID) string {
	return id.String()
}

// string2UUID converts string to uuid.UUID type
func string2UUID(id string) (uuid.UUID, error) {
	r, err := uuid.Parse(id)
	if err != nil {
		return [16]byte{}, err
	}
	return r, nil
}

// uuidPtrToNullable converts an optional string UUID pointer to pgtype.UUID.
func uuidPtr2Nullable(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: u, Valid: true}
}

func nUUIDPtr2Nullable(s *string) pgtype.UUID {
	if s == nil {
		return pgtype.UUID{Valid: false}
	}
	return uuidPtr2Nullable(*s)
}

// stringPtrToNullable safely converts string pointer to pgtype.Text.
func stringPtrToNullable(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return stringToNullable(*s)
}

func stringToNullable(s string) pgtype.Text {
	if len(s) == 0 || s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}
