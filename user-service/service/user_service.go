package service

import (
	"context"
	"errors"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "user-service/auth-pb"
	"user-service/dto"
	"user-service/models"
	"user-service/repository"
)

type UserService interface {
	RegisterUser(ctx context.Context, data *dto.RegisterUserDto) (string, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	AssignRole(ctx context.Context, userId string, roleName string) error
	DeleteUserById(ctx context.Context, id string) error
}

type userService struct {
	authClient  authpb.AuthServiceClient
	repository  repository.UserRepository
	roleService RoleService
}

func NewUserService(repository repository.UserRepository, roleService RoleService, authClient authpb.AuthServiceClient) *userService {
	return &userService{
		repository:  repository,
		roleService: roleService,
		authClient:  authClient,
	}
}

func (s *userService) RegisterUser(ctx context.Context, data *dto.RegisterUserDto) (string, error) {
	var userId string

	createUserDto := dto.CreateUserDto{
		Name:     data.Name,
		Username: data.Username,
	}

	registerCredentialsDto := dto.RegisterCredentialsDto{
		Username: data.Username,
		Password: data.Password,
	}

	err := s.repository.InTransaction(ctx, func(txRepo repository.UserRepository) error {
		var txErr error

		userId, txErr = txRepo.CreateUser(ctx, &createUserDto)
		if txErr != nil {
			log.Printf("failed to create user: %v", txErr)
			return status.Errorf(codes.Internal, "failed to create user.")
		}

		registerCredentialsDto.Id = userId
		if txErr = registerCredentials(ctx, s.authClient, registerCredentialsDto); txErr != nil {
			log.Printf("failed to register credentials: %v", txErr)
			return status.Errorf(codes.Internal, "failed to register credentials.")
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return userId, nil
}

func (s *userService) GetUserById(ctx context.Context, id string) (*models.User, error) {
	user, err := s.repository.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrEntityNotFound) {
			return nil, status.Error(codes.NotFound, "User not found.")
		}
		log.Printf("failed to get user by id: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get user.")
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrEntityNotFound) {
			return nil, status.Error(codes.NotFound, "User not found.")
		}
		log.Printf("failed to get user by username: %v", err)
		return nil, status.Error(codes.Internal, "Failed to get user.")
	}
	return user, nil
}

func (s *userService) AssignRole(ctx context.Context, userId string, roleName string) error {
	role, err := s.roleService.GetRoleByName(ctx, roleName)
	if err != nil {
		return err
	}

	if err = s.repository.AssignRole(ctx, userId, role); err != nil {
		if errors.Is(err, repository.ErrEntityNotFound) {
			return status.Error(codes.NotFound, "user not found.")
		}
		log.Printf("failed to assign role to user: %v", err)
		return status.Error(codes.Internal, "failed to assign role to user.")
	}

	return err
}

func (s *userService) DeleteUserById(ctx context.Context, id string) error {
	if err := s.repository.DeleteUserById(ctx, id); err != nil {
		if errors.Is(err, repository.ErrEntityNotFound) {
			return status.Error(codes.NotFound, "User not found.")
		}
		log.Printf("failed to delete user: %v", err)
		return status.Error(codes.Internal, "Failed to delete user.")
	}
	return nil
}

func registerCredentials(ctx context.Context, authClient authpb.AuthServiceClient, dto dto.RegisterCredentialsDto) error {
	_, err := authClient.RegisterCredentials(ctx, &authpb.RegisterCredentialsRequest{
		UserId:   dto.Id,
		Username: dto.Username,
		Password: dto.Password,
	})
	if err != nil {
		log.Printf("failed to register credentials: %v", err)
		return err
	}
	return nil
}
