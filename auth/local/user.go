package local

import "github.com/google/uuid"

type AuthUser struct {
	ID           uuid.UUID
	Username     string
	PasswordHash []byte
}
