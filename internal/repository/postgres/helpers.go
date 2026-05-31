package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// pgTimeTZ extracts time.Time from a pgtype.Timestamptz value.
// Falls back to a zero value if the timestamp is not valid.
func pgTimeTZ(ts pgtype.Timestamptz, _ *time.Location) time.Time {
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
func string2UUID(id string) uuid.UUID {
	return uuid.MustParse(id)
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

// usernamePtrToNullable safely converts string pointer to pgtype.Text.
func usernamePtrToNullable(s *string) pgtype.Text {
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

func nFloat2Nullable(f *float64) pgtype.Float8 {
	if f == nil {
		return pgtype.Float8{Valid: false}
	}
	return float2Nullable(*f)
}

func float2Nullable(f float64) pgtype.Float8 {
	if f == 0 {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: f, Valid: true}
}
