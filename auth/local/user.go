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

func (u *AuthUser) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password))
}

func (u *AuthUser) SetPassword(password string) error {
	// note: this should probably be even stricter :)
	if len(password) <= 8 {
		return ErrWeakPassword
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = passwordBytes
	return nil
}
