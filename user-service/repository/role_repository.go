package repository

import (
	"context"
	"errors"
	"user-service/dto"
	"user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository interface {
	CreateRole(ctx context.Context, data *dto.CreateRoleDto) (*models.Role, error)
	CreateRoles(ctx context.Context, data []dto.CreateRoleDto) ([]models.Role, error)
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	GetRolesByNames(ctx context.Context, name []string) ([]models.Role, error)
	DeleteRoleById(ctx context.Context, id string) error
}

type gormRoleRepository struct {
	db *gorm.DB
}

func NewGormRoleRepository(db *gorm.DB) *gormRoleRepository {
	return &gormRoleRepository{db: db}
}

func (r *gormRoleRepository) CreateRole(ctx context.Context, data *dto.CreateRoleDto) (*models.Role, error) {
	role := &models.Role{
		ID:   uuid.NewString(),
		Name: data.Name,
	}
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateKey
		}
		return nil, err
	}
	return role, nil
}

func (r *gormRoleRepository) CreateRoles(ctx context.Context, data []dto.CreateRoleDto) ([]models.Role, error) {
	roles := make([]models.Role, len(data))
	for i := range data {
		roles[i] = models.Role{
			ID:   uuid.NewString(),
			Name: data[i].Name,
		}
	}
	if err := r.db.WithContext(ctx).Create(&roles).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrDuplicateKey
		}
		return nil, err
	}
	return roles, nil
}

func (r *gormRoleRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	role := &models.Role{Name: name}
	if err := r.getRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (r *gormRoleRepository) GetRolesByNames(ctx context.Context, names []string) ([]models.Role, error) {
	roles := make([]models.Role, len(names))
	for i, name := range names {
		roles[i] = models.Role{Name: name}
	}
	if err := r.getRoles(ctx, &roles); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *gormRoleRepository) DeleteRoleById(ctx context.Context, id string) error {
	role := &models.Role{ID: id}
	return r.deleteRole(ctx, role)
}

func (r *gormRoleRepository) getRole(ctx context.Context, role *models.Role) error {
	if err := r.db.WithContext(ctx).Where(role).First(role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEntityNotFound
		}
		return err
	}
	return nil
}

func (r *gormRoleRepository) getRoles(ctx context.Context, roles *[]models.Role) error {
	if err := r.db.WithContext(ctx).Where(roles).Find(roles).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEntityNotFound
		}
		return err
	}
	return nil
}

func (r *gormRoleRepository) deleteRole(ctx context.Context, role *models.Role) error {
	result := r.db.WithContext(ctx).Delete(&role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEntityNotFound
	}
	return nil
}
