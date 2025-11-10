package dto

import "time"

type SaveRefreshToken struct {
	RefreshToken string
	UserID       string
	Expiration   time.Time
}
