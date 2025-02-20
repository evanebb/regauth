package pat

import (
	"github.com/google/uuid"
	"net"
	"time"
)

type PersonalAccessToken struct {
	ID             uuid.UUID
	Hash           []byte
	Description    string
	Permission     Permission
	ExpirationDate time.Time
	UserID         uuid.UUID
}

type Permission string

const (
	PermissionReadOnly        Permission = "readOnly"
	PermissionReadWrite       Permission = "readWrite"
	PermissionReadWriteDelete Permission = "readWriteDelete"
)

func (t Permission) GetAllowedActions() []string {
	m := map[Permission][]string{
		PermissionReadOnly:        {"pull"},
		PermissionReadWrite:       {"pull", "push"},
		PermissionReadWriteDelete: {"pull", "push", "delete"},
	}

	a, ok := m[t]
	if !ok {
		return []string{}
	}
	return a
}

func (t Permission) HumanReadable() string {
	m := map[Permission]string{
		PermissionReadOnly:        "Read-only",
		PermissionReadWrite:       "Read and write",
		PermissionReadWriteDelete: "Read, write and delete",
	}

	h, ok := m[t]
	if !ok {
		return ""
	}
	return h
}

// UsageLogEntry is a single entry in the usage log of a PersonalAccessToken.
type UsageLogEntry struct {
	TokenID   uuid.UUID
	SourceIP  net.IP
	Timestamp time.Time
}
