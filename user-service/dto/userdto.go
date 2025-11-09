package dto

import "user-service/models"

type CreateUserDto struct {
	Name     string
	Username string
	Password string
}

type UpdateUserDto struct {
	Name  *string
	Roles *[]models.Role
}
