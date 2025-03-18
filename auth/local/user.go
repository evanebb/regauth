package local

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUser struct {
	ID           uuid.UUID
	Username     string
	PasswordHash []byte
}

func (u AuthUser) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
}
