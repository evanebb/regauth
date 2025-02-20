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
	PermissionType PermissionType
	ExpirationDate time.Time
	UserID         uuid.UUID
}

type PermissionType string

const (
	ReadOnly        PermissionType = "readOnly"
	ReadWrite       PermissionType = "readWrite"
	ReadWriteDelete PermissionType = "readWriteDelete"
)

func (t PermissionType) GetAllowedActions() []string {
	m := map[PermissionType][]string{
		ReadOnly:        {"pull"},
		ReadWrite:       {"pull", "push"},
		ReadWriteDelete: {"pull", "push", "delete"},
	}

	a, ok := m[t]
	if !ok {
		return []string{}
	}
	return a
}

func (t PermissionType) HumanReadable() string {
	m := map[PermissionType]string{
		ReadOnly:        "Read-only",
		ReadWrite:       "Read and write",
		ReadWriteDelete: "Read, write and delete",
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
