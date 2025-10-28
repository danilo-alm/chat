package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	authpb "user-service/auth-pb"
	"user-service/dto"
	"user-service/models"
	pb "user-service/pb"
	"user-service/repository"
	"user-service/server"
	"user-service/service"
	"user-service/utils"
)

func main() {
	var err error

	authServiceUrl := utils.GetEnv("AUTH_SERVICE_URL", "auth-service:50051")
	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	authClient := authpb.NewAuthServiceClient(conn)
	defer conn.Close()

	db_uri := utils.GetEnv("MARIADB_URI", "user:secret@tcp(user-db:3306)/userdb")
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", db_uri)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{})

	var adminRole = models.Role{Name: "ADMIN"}
	var adminUser = models.User{
		Name:     "Admin",
		Username: "admin",
		Roles:    []models.Role{adminRole},
	}

	if err := Seed(db, &adminRole); err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	port := utils.GetEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	userRepository := repository.NewGormUserRepository(db)
	userService := service.NewUserService(authClient, userRepository)
	server := server.NewUserServer(userService)

	adminPassword := utils.GetEnv("ADMIN_PASSWORD", "admin")
	_, err = userService.RegisterUser(context.Background(), &dto.RegisterUserDto{
		Name:     adminUser.Name,
		Username: adminUser.Username,
	})
	if err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userServer)

	enableReflection := utils.GetEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enableReflection)
	if enableReflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
