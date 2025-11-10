package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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
