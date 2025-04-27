package global

import "github.com/google/uuid"

func NewUUID() *uuid.UUID {
	u, _ := uuid.NewV7()
	return &u
}

func GetCallID() string {
	uid := NewUUID()
	return uid.String()
}

func GetViaBranch() string {
	uid := NewUUID()
	return MagicCookie + uid.String()[24:]
}

func GetTagOrKey() string {
	uid := NewUUID()
	return uid.String()[24:]
}
