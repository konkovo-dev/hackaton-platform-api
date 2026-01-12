package auth

import "time"

type Claims struct {
	UserID    string
	ExpiresAt time.Time
}
