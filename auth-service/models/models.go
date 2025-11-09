package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)"`
	Token     string    `gorm:"not null;uniqueIndex;type:varchar(255)"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (rt *RefreshToken) BeforeCreate(tx gorm.DB) error {
	if rt.ID == "" {
		rt.ID = uuid.NewString()
	}
	return nil
}
