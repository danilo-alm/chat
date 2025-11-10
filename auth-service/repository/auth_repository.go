package repository

import (
	"auth-service/dto"
	"auth-service/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type AuthRepository interface {
	SaveRefreshToken(ctx context.Context, data *dto.SaveRefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error)
	DeleteRefreshTokenById(ctx context.Context, id string) error
	RotateRefreshToken(ctx context.Context, oldToken string, newToken *dto.SaveRefreshToken) error
}

type gormAuthRepository struct {
	db *gorm.DB
}

func NewGormAuthRepository(db *gorm.DB) AuthRepository {
	return &gormAuthRepository{db: db}
}

func (r *gormAuthRepository) SaveRefreshToken(ctx context.Context, data *dto.SaveRefreshToken) error {
	rt := &models.RefreshToken{
		Token:     data.RefreshToken,
		ExpiresAt: data.Expiration,
	}

	err := r.db.Create(&rt).Error

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrDuplicateKey
	}

	return err
}

func (r *gormAuthRepository) GetRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	rt := &models.RefreshToken{Token: token}
	if err := r.db.WithContext(ctx).Where(&rt).First(&rt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}
	return rt, nil
}

func (r *gormAuthRepository) DeleteRefreshTokenById(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("ID = ?", id).Delete(&models.RefreshToken{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (r *gormAuthRepository) RotateRefreshToken(ctx context.Context, oldToken string, newToken *dto.SaveRefreshToken) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("token = ?", oldToken).Delete(&models.RefreshToken{})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return ErrEntityNotFound
		}

		rt := &models.RefreshToken{
			Token:     newToken.RefreshToken,
			UserID:    newToken.UserID,
			ExpiresAt: newToken.Expiration,
		}

		if err := tx.Create(&rt).Error; err != nil {
			return err
		}

		return nil
	})
}
