package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	authpb "user-service/auth-pb"
	pb "user-service/pb"
)

type User struct {
	Id       string `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Username string `gorm:"not null;uniqueIndex"`
}

type server struct {
	pb.UnimplementedUserServiceServer
	authClient    authpb.AuthServiceClient
	mariadbClient *gorm.DB
}

func main() {
	var err error

	authServiceUrl := getEnv("AUTH_SERVICE_URL", "auth-service:50051")
	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Auth Service: %v", err)
	}
	defer conn.Close()

	db_uri := getEnv("MARIADB_URI", "user:secret@tcp(user-db:3306)/userdb")
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", db_uri)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&User{})

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	userServer := &server{
		authClient:    authpb.NewAuthServiceClient(conn),
		mariadbClient: db,
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userServer)

	enableReflection := getEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enableReflection)
	if enableReflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	var resp *pb.CreateUserResponse

	err := s.mariadbClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := User{
			Id:       uuid.New().String(),
			Name:     req.GetName(),
			Username: req.GetUsername(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		_, err := s.authClient.RegisterCredentials(ctx, &authpb.RegisterCredentialsRequest{
			UserId:   user.Id,
			Password: req.GetPassword(),
		})
		if err != nil {
			return fmt.Errorf("failed to register credentials: %w", err)
		}

		resp = &pb.CreateUserResponse{Id: user.Id}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserResponse, error) {
	var user User
	if err := s.mariadbClient.WithContext(ctx).First(&user, "id = ?", req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &pb.GetUserResponse{
		User: mapUserToPbUser(user),
	}, nil
}

func (s *server) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserResponse, error) {
	var user User
	if err := s.mariadbClient.WithContext(ctx).First(&user, "username = ?", req.GetUsername()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &pb.GetUserResponse{
		User: mapUserToPbUser(user),
	}, nil
}

func getUser(ctx context.Context, c *mongo.Client, filter any) (*pb.GetUserResponse, error) {
	// ...removed: now handled in GetUserById and GetUserByUsername...
	return nil, fmt.Errorf("not implemented")
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	result := s.mariadbClient.WithContext(ctx).Delete(&User{}, "id = ?", req.GetId())
	if result.Error != nil {
		return nil, fmt.Errorf("database error: %v", result.Error)
	}
	return &pb.DeleteUserResponse{
		Success: result.RowsAffected == 1,
	}, nil
}

func mapUserToPbUser(user User) *pb.User {
	return &pb.User{
		Id:       user.Id,
		Username: user.Username,
		Name:     user.Name,
	}
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
