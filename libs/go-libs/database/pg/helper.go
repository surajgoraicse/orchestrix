package pg

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToTimePtr[T pgtype.Timestamp | pgtype.Timestamptz](ts T) *time.Time {
	switch v := any(ts).(type) {
	case pgtype.Timestamp:
		if !v.Valid {
			return nil
		}
		t := v.Time
		return &t
	case pgtype.Timestamptz:
		if !v.Valid {
			return nil
		}
		t := v.Time
		return &t
	default:
		return nil
	}
}

func ToStringPtr(txt pgtype.Text) *string {
	if !txt.Valid {
		return nil
	}
	s := txt.String
	return &s
}

func GetPgUUIDFromUUID(uuid uuid.UUID) (pgtype.UUID, error) {
	b, err := uuid.MarshalBinary()
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{
		Bytes: [16]byte(b),
		Valid: true,
	}, nil
}
