package utils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

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
