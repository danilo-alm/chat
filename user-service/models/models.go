package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       string `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Username string `gorm:"not null;uniqueIndex"`
	Password string `gorm:"not null"`
	Roles    []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
	ID   string `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex;not null"`
}

type UserRole struct {
	UserID string `gorm:"primaryKey"`
	RoleID string `gorm:"primaryKey"`
	User   User
	Role   Role
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	setIDIfEmpty(&u.ID)
	return nil
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	setIDIfEmpty(&r.ID)
	return nil
}

func setIDIfEmpty(id *string) {
	if *id == "" {
		*id = uuid.NewString()
	}
}
