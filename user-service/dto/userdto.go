package dto

type RegisterUserDto struct {
	Name     string
	Username string
	Password string
}

type CreateUserDto struct {
	Name     string
	Username string
}

type RegisterCredentialsDto struct {
	Id       string
	Username string
	Password string
}
