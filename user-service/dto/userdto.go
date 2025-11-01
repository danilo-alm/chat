package dto

import "user-service/models"

type RegisterUserDto struct {
	Name     string
	Username string
	Password string
}

type CreateUserDto struct {
	Name     string
	Username string
}

type UpdateUserDto struct {
	Name  *string
	Roles *[]models.Role
}

type RegisterCredentialsDto struct {
	Id       string
	Username string
	Password string
}
