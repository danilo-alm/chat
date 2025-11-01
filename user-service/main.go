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

	authpb "user-service/auth-pb"
	"user-service/models"
	pb "user-service/pb"
	"user-service/repository"
	"user-service/server"
	"user-service/service"
	"user-service/utils"
)

func main() {
	db, err := initializeDatabase()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	authClient, conn, err := initializeAuthClient()
	if err != nil {
		log.Fatalf("Auth client initialization failed: %v", err)
	}
	defer conn.Close()

	roleRepository := repository.NewGormRoleRepository(db)
	roleService := service.NewRoleService(roleRepository)

	userRepository := repository.NewGormUserRepository(db)
	userService := service.NewUserService(userRepository, roleService, authClient)

	if err := SeedAdminUser(db, authClient); err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	serverPort := utils.GetEnv("GRPC_PORT", "50051")
	lis, s, err := setupGRPCServer(serverPort, userService, roleService)
	if err != nil {
		log.Fatalf("gRPC server setup failed: %v", err)
	}

	log.Println("gRPC server started on port", serverPort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initializeDatabase() (*gorm.DB, error) {
	mariadbURI := utils.GetEnv("MARIADB_URI", "user:secret@tcp(user-db:3306)/userdb")
	uriWithOptions := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", mariadbURI)

	db, err := gorm.Open(mysql.Open(uriWithOptions), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Role{}, &models.UserRole{})
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

func initializeAuthClient() (authpb.AuthServiceClient, *grpc.ClientConn, error) {
	authServiceUrl := utils.GetEnv("AUTH_SERVICE_URL", "auth-service:50051")

	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to Auth Service at %s: %w", authServiceUrl, err)
	}

	authClient := authpb.NewAuthServiceClient(conn)
	return authClient, conn, nil
}

func setupGRPCServer(port string, userService service.UserService, roleService service.RoleService) (net.Listener, *grpc.Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	userServerImpl := server.NewUserServer(userService, roleService)
	roleServerImpl := server.NewRoleServer(roleService)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userServerImpl)
	pb.RegisterRoleServiceServer(s, roleServerImpl)

	enableReflection := utils.GetEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enableReflection)
	if enableReflection == "true" {
		reflection.Register(s)
	}

	return lis, s, nil
}
