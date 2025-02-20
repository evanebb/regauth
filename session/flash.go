package session

import "encoding/gob"

type Flash struct {
	Type    FlashType
	Message string
}

func NewFlash(t FlashType, m string) Flash {
	return Flash{Type: t, Message: m}
}

type FlashType string

var (
	FlashTypeSuccess FlashType = "success"
	FlashTypeWarning FlashType = "warning"
	FlashTypeError   FlashType = "error"
)

func init() {
	// Necessary for this to work with gorilla/sessions
	gob.Register(&Flash{})
}
