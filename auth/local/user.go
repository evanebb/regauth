package local

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserCredentials struct {
	UserID       uuid.UUID
	PasswordHash []byte
}

func (c *UserCredentials) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword(c.PasswordHash, []byte(password))
}

func (c *UserCredentials) SetPassword(password string) error {
	// note: this should probably be even stricter :)
	if len(password) <= 8 {
		return ErrWeakPassword
	}

	passwordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	c.PasswordHash = passwordBytes
	return nil
}
