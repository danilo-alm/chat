package main

import (
	"context"
	"fmt"
	"user-service/models"
	"user-service/utils"

	authpb "user-service/auth-pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func SeedAdminUser(db *gorm.DB, authClient authpb.AuthServiceClient) error {
	adminRole := models.Role{}

	if err := db.FirstOrCreate(&adminRole, models.Role{Name: "ADMIN"}).Error; err != nil {
		return fmt.Errorf("failed to seed admin role: %w", err)
	}

	adminUser := models.User{
		Name:     "Admin",
		Username: "admin",
		Roles:    []models.Role{adminRole},
	}
	adminPassword := utils.GetEnv("ADMIN_PASSWORD", "admin")

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.FirstOrCreate(&adminUser, models.User{Username: "admin"}).Error; err != nil {
			return fmt.Errorf("failed to seed admin user: %w", err)
		}

		_, err := authClient.RegisterCredentials(context.Background(), &authpb.RegisterCredentialsRequest{
			UserId:   adminUser.ID,
			Username: adminUser.Username,
			Password: adminPassword,
		})
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() != codes.AlreadyExists {
				return fmt.Errorf("failed to register admin user credentials: %w", err)
			}
		}

		return nil
	})
}
