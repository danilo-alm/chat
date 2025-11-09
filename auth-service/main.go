package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"auth-service/config"
	"auth-service/models"
	pb "auth-service/pb"
	"auth-service/repository"
	"auth-service/server"
	"auth-service/service"
	userpb "auth-service/user-pb"
	"auth-service/utils"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	userServiceConn, err := connectToUserService()
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer userServiceConn.Close()

	userServiceClient := userpb.NewUserServiceClient(userServiceConn)

	authRepository := repository.NewGormAuthRepository(db)
	authService := service.NewAuthService(authRepository, userServiceClient, cfg)

	serverPort := utils.GetEnv("GRPC_PORT", "50051")
	lis, s, err := setupGRPCServer(serverPort, authService)
	if err != nil {
		log.Fatalf("gRPC server setup failed: %v", err)
	}

	log.Println("gRPC server started on port", serverPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initializeDatabase() (*gorm.DB, error) {
	mariadbURI := utils.GetEnv("MARIADB_URI", "user:secret@tcp(auth-db:3306)/authdb")
	uriWithOptions := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", mariadbURI)

	db, err := gorm.Open(mysql.Open(uriWithOptions), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(&models.RefreshToken{})
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

func connectToUserService() (*grpc.ClientConn, error) {
	userServiceAddr := utils.GetEnv("USER_SERVICE_URL", "user-service:50051")
	conn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to User Service: %w", err)
	}
	return conn, nil
}

func setupGRPCServer(port string, authService service.AuthService) (net.Listener, *grpc.Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	authServer := server.NewAuthServer(authService)

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, authServer)

	enableReflection := utils.GetEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enableReflection)
	if enableReflection == "true" {
		reflection.Register(s)
	}

	return lis, s, nil
}
