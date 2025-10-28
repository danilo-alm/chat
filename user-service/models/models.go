package models

type User struct {
	ID       string `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Username string `gorm:"not null;uniqueIndex"`
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

type Identifiable interface {
	SetID(id string)
}

func (u *User) SetID(id string) {
	u.ID = id
}

func (u *Role) SetID(id string) {
	u.ID = id
}
