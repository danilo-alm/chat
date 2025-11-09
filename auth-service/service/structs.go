package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type refreshToken struct {
	Token     string    `gorm:"primaryKey;type:varchar(255)"`
	ExpiresAt time.Time `gorm:"column:expires_at;index"`
	UserId    string    `gorm:"column:user_id;type:varchar(36);index"`
}

type tokenConfig struct {
	accessTTL     time.Duration
	refreshTTL    time.Duration
	accessSecret  []byte
	refreshSecret []byte
}

type Tokens struct {
	Access     string
	AccessExp  time.Time
	Refresh    string
	RefreshExp time.Time
}

type claims struct {
	userId   string
	username string
	roles    []string
}

type jwtClaims struct {
	jwt.RegisteredClaims

	Username string
	Roles    []string
}
