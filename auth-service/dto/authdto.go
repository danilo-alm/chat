package dto

import "time"

type SaveRefreshToken struct {
	RefreshToken string
	Expiration   time.Time
}
