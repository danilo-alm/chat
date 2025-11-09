package repository

import (
	"context"
	"errors"
	"user-service/dto"
	"user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, data *dto.CreateUserDto) (string, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	AssignRole(ctx context.Context, userId string, role *models.Role) error
	DeleteUserById(ctx context.Context, id string) error
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) CreateUser(ctx context.Context, data *dto.CreateUserDto) (string, error) {
	user := &models.User{
		ID:       uuid.NewString(),
		Name:     data.Name,
		Username: data.Username,
		Password: data.Password,
	}
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return "", ErrDuplicateKey
		}
		return "", err
	}
	return user.ID, nil
}

func (r *gormUserRepository) GetUserById(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{ID: id}
	if err := r.getUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *gormUserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{Username: username}
	if err := r.getUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r *gormUserRepository) AssignRole(ctx context.Context, userId string, role *models.Role) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := &models.User{ID: userId}
		if err := tx.Preload("Roles").Where(user).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEntityNotFound
			}
			return err
		}

		return tx.Model(user).Association("Roles").Append(role)
	})
}

func (r *gormUserRepository) UpdateUserById(ctx context.Context, id string, data *dto.UpdateUserDto) (*models.User, error) {
	var user *models.User

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user = &models.User{ID: id}
		if err := tx.Preload("Roles").Where(user).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEntityNotFound
			}
			return err
		}

		if data.Name != nil {
			user.Name = *data.Name
		}

		if err := tx.Save(user).Error; err != nil {
			return err
		}

		if data.Roles != nil {
			if err := tx.Model(user).Association("Roles").Replace(data.Roles); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *gormUserRepository) DeleteUserById(ctx context.Context, id string) error {
	user := &models.User{ID: id}
	if err := r.deleteUser(ctx, user); err != nil {
		return err
	}
	return nil
}

func (r *gormUserRepository) getUser(ctx context.Context, user *models.User) error {
	err := r.db.WithContext(ctx).Preload("Roles").Where(&user).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrEntityNotFound
	}
	return err
}

func (r *gormUserRepository) deleteUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Delete(&user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrEntityNotFound
	}
	return nil
}
