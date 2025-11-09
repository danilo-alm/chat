package service

import (
	"context"
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"user-service/dto"
	"user-service/models"
	"user-service/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, data *dto.CreateUserDto) (string, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	AssignRole(ctx context.Context, userId string, roleName string) error
	DeleteUserById(ctx context.Context, id string) error
}

type userService struct {
	repository  repository.UserRepository
	roleService RoleService
}

func NewUserService(repository repository.UserRepository, roleService RoleService) UserService {
	return &userService{
		repository:  repository,
		roleService: roleService,
	}
}

func (s *userService) RegisterUser(ctx context.Context, data *dto.CreateUserDto) (string, error) {
	genericError := status.Error(codes.Internal, "failed to create user.")

	hashedPassword, err := hashPassword(data.Password)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return "", genericError
	}

	userId, err := s.repository.CreateUser(ctx, &dto.CreateUserDto{
		Name:     data.Name,
		Username: data.Username,
		Password: hashedPassword,
	})
	if errors.Is(err, repository.ErrDuplicateKey) {
		return "", status.Error(codes.AlreadyExists, "username already exists.")
	} else if err != nil {
		log.Printf("failed to create user: %v", err)
		return "", genericError
	}

	return userId, nil
}

func (s *userService) GetUserById(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repository.GetUserById(ctx, id)
	return handleFetchedUser(user, err)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	return handleFetchedUser(user, err)
}

func (s *userService) AssignRole(ctx context.Context, userId string, roleName string) error {
	role, err := s.roleService.GetRoleByName(ctx, roleName)
	if err != nil {
		return err
	}

	err = s.repository.AssignRole(ctx, userId, role)
	if errors.Is(err, repository.ErrEntityNotFound) {
		return status.Error(codes.NotFound, "user not found.")
	} else if err != nil {
		log.Printf("failed to assign role to user: %v", err)
		return status.Error(codes.Internal, "failed to assign role to user.")
	}

	return err
}

func (s *userService) DeleteUserById(ctx context.Context, id string) error {
	err := s.repository.DeleteUserById(ctx, id)

	if errors.Is(err, repository.ErrEntityNotFound) {
		return status.Error(codes.NotFound, "user not found.")
	} else if err != nil {
		log.Printf("failed to delete user: %v", err)
		return status.Error(codes.Internal, "failed to delete user.")
	}

	return nil
}

func handleFetchedUser(user *models.User, err error) (*models.User, error) {
	if errors.Is(err, repository.ErrEntityNotFound) {
		return nil, status.Error(codes.NotFound, "User not found.")
	} else if err != nil {
		log.Printf("failed to get user: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get user.")
	}
	return user, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
