package models

import (
	"time"
)

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:varchar(36);default:REPLACE(UUID(),'-','')"`
	Token     string    `gorm:"not null;uniqueIndex;type:varchar(36)"`
	UserID    string    `gorm:"not null;type:varchar(36);index"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
