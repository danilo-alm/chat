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
	DeleteUserById(ctx context.Context, id string) error
}

type userService struct {
	authClient authpb.AuthServiceClient
	repository repository.UserRepository
}

func NewUserService(authClient authpb.AuthServiceClient, repository repository.UserRepository) *userService {
	return &userService{
		authClient: authClient,
		repository: repository,
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
			return status.Errorf(codes.Internal, "Failed to create user.")
		}

		registerCredentialsDto.Id = userId
		if txErr = registerCredentials(ctx, s.authClient, registerCredentialsDto); txErr != nil {
			log.Printf("failed to register credentials: %v", txErr)
			return status.Errorf(codes.Internal, "Failed to register credentials.")
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
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) DeleteUserById(ctx context.Context, id string) error {
	if err := s.repository.DeleteUserById(ctx, id); err != nil {
		log.Printf("failed to delete user: %v", err)
		return err
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
		return errors.New("failed to register credentials")
	}
	return nil
}
