package main

import (
	"fmt"
	"user-service/models"
	"user-service/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdmin(db *gorm.DB) error {
	adminRole := models.Role{Name: "ADMIN"}

	if err := db.FirstOrCreate(&adminRole, models.Role{Name: "ADMIN"}).Error; err != nil {
		return fmt.Errorf("failed to seed admin role: %w", err)
	}

	adminPassword := utils.GetEnv("ADMIN_PASSWORD", "admin")
	hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	adminUser := models.User{
		Username: "admin",
		Name:     "Admin",
		Password: string(hashedAdminPassword),
		Roles:    []models.Role{adminRole},
	}

	if err := db.FirstOrCreate(&adminUser, models.User{Username: "admin"}).Error; err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	return nil
}
