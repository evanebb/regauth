package token

import (
	"github.com/google/uuid"
	"net"
	"regexp"
	"time"
)

type PersonalAccessToken struct {
	ID             uuid.UUID   `json:"id"`
	Description    Description `json:"description"`
	Permission     Permission  `json:"permission"`
	ExpirationDate time.Time   `json:"expirationDate"`
	UserID         uuid.UUID   `json:"-"`
	CreatedAt      time.Time   `json:"createdAt"`
}

func (t PersonalAccessToken) IsValid() error {
	if err := t.Permission.IsValid(); err != nil {
		return err
	}

	if err := t.Description.IsValid(); err != nil {
		return err
	}

	return nil
}

type Description string

var validDescription = regexp.MustCompile(`^[a-zA-Z0-9-_ ]+$`)

func (d Description) IsValid() error {
	l := len([]rune(d))
	if l < 2 {
		return InvalidDescriptionError("description cannot be shorter than 2 characters")
	}

	if l > 255 {
		return InvalidDescriptionError("description cannot be longer than 255 characters")
	}

	if !validDescription.MatchString(string(d)) {
		return InvalidDescriptionError(`description can only contain alphanumeric characters, spaces, "-" and "_"`)
	}

	return nil
}

type Permission string

const (
	PermissionReadOnly        = Permission("readOnly")
	PermissionReadWrite       = Permission("readWrite")
	PermissionReadWriteDelete = Permission("readWriteDelete")
)

func (p Permission) IsValid() error {
	if p != PermissionReadOnly && p != PermissionReadWrite && p != PermissionReadWriteDelete {
		return ErrInvalidPermission
	}

	return nil
}

func (p Permission) GetAllowedActions() []string {
	m := map[Permission][]string{
		PermissionReadOnly:        {"pull"},
		PermissionReadWrite:       {"pull", "push"},
		PermissionReadWriteDelete: {"pull", "push", "delete"},
	}

	a, ok := m[p]
	if !ok {
		return []string{}
	}
	return a
}

// UsageLogEntry is a single entry in the usage log of a PersonalAccessToken.
type UsageLogEntry struct {
	TokenID   uuid.UUID
	SourceIP  net.IP
	Timestamp time.Time
}
