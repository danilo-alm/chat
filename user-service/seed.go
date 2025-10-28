package main

import (
	"errors"
	"user-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func Seed[T models.Identifiable](db *gorm.DB, entity T) error {
	if err := db.Where(entity).First(entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			entity.SetID(uuid.NewString())
			if err := db.Create(entity).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
