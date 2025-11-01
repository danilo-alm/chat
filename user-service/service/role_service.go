package service

import (
	"context"
	"errors"
	"log"
	"user-service/dto"
	"user-service/models"
	"user-service/repository"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RoleService interface {
	CreateRole(ctx context.Context, role *dto.CreateRoleDto) (*models.Role, error)
	CreateRoles(ctx context.Context, roles []dto.CreateRoleDto) ([]models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	GetRolesByNames(ctx context.Context, names []string) ([]models.Role, error)
	DeleteRoleById(ctx context.Context, id string) error
}

type roleService struct {
	repository repository.RoleRepository
}

func NewRoleService(repository repository.RoleRepository) *roleService {
	return &roleService{
		repository: repository,
	}
}

func (s *roleService) CreateRole(ctx context.Context, data *dto.CreateRoleDto) (*models.Role, error) {
	roles, err := s.repository.CreateRole(ctx, data)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return nil, status.Errorf(codes.AlreadyExists, "Role with name %s already exists", data.Name)
		}
		log.Printf("failed to create role: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create role")
	}
	return roles, err
}

func (s *roleService) CreateRoles(ctx context.Context, data []dto.CreateRoleDto) ([]models.Role, error) {
	roles, err := s.repository.CreateRoles(ctx, data)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			return nil, status.Errorf(codes.AlreadyExists, "One or more roles already exist")
		}
		log.Printf("failed to create roles: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to create roles")
	}
	return roles, nil
}

func (s *roleService) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	role, err := s.repository.GetRoleByName(ctx, name)
	if err != nil {
		log.Printf("failed to get role by name: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get role by name")
	}
	return role, nil
}

func (s *roleService) GetRolesByNames(ctx context.Context, names []string) ([]models.Role, error) {
	roles, err := s.repository.GetRolesByNames(ctx, names)
	if err != nil {
		log.Printf("failed to get roles by names: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get roles by names")
	}
	return roles, nil
}

func (s *roleService) DeleteRoleById(ctx context.Context, id string) error {
	if err := s.repository.DeleteRoleById(ctx, id); err != nil {
		log.Printf("failed to delete role by id: %v", err)
		return status.Errorf(codes.Internal, "failed to delete role by id")
	}
	return nil
}
