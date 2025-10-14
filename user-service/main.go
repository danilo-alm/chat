package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("username already exists (constraint violation): %v", err)
				return status.Errorf(codes.AlreadyExists, "user already exists")
			}
			log.Printf("failed to create user: %v", err)
			return status.Errorf(codes.Internal, "failed to create user")
		}

		_, err := s.authClient.RegisterCredentials(ctx, &authpb.RegisterCredentialsRequest{
			UserId:   user.Id,
			Username: user.Username,
			Password: req.GetPassword(),
		})
		if err != nil {
			log.Printf("failed to register credentials: %v", err)
			return status.Errorf(codes.Internal, "failed to register credentials")
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
	return getUser(ctx, s.mariadbClient, User{Id: req.GetId()})
}

func (s *server) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserResponse, error) {
	return getUser(ctx, s.mariadbClient, User{Username: req.GetUsername()})
}

func getUser(ctx context.Context, mariadbClient *gorm.DB, user User) (*pb.GetUserResponse, error) {
	if err := mariadbClient.WithContext(ctx).Where(&user).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("user not found: %v", err)
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		log.Printf("database error: %v", err)
		return nil, status.Errorf(codes.Internal, "could not retrieve user")
	}
	return &pb.GetUserResponse{
		User: mapUserToPbUser(user),
	}, nil
}

func (s *server) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.mariadbClient.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := s.mariadbClient.WithContext(ctx).Delete(&User{Id: req.GetId()})
		if result.Error != nil {
			log.Printf("database error: %v", result.Error)
			return status.Errorf(codes.Internal, "failed to delete user")
		}
		if result.RowsAffected == 0 {
			log.Printf("could not delete user (not found): %s", req.GetId())
			return status.Errorf(codes.NotFound, "user not found")
		}

		_, err := s.authClient.DeleteCredentials(ctx, &authpb.DeleteCredentialsRequest{
			UserId: req.GetId(),
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
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
