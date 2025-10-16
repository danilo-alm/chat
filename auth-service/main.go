package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	pb "auth-service/pb"
)

type server struct {
	pb.UnimplementedAuthServiceServer
	mariadbClient *gorm.DB
}

type Credentials struct {
	UserId   string         `gorm:"primaryKey"`
	Username string         `gorm:"unique;not null;index"`
	Password string         `gorm:"not null"`
	Tokens   []RefreshToken `gorm:"foreignKey:UserId;references:UserId;constraint:OnDelete:CASCADE"`
}

type RefreshToken struct {
	Id        string    `gorm:"primaryKey"`
	UserId    string    `gorm:"not null;index"`
	Token     string    `gorm:"not null;uniqueIndex;size:512"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	ExpiresAt time.Time `gorm:"not null"`
}

const (
	accessTTL  = 24 * time.Hour
	refreshTTL = 7 * 24 * time.Hour
)

var (
	accessSecret  = []byte(os.Getenv("ACCESS_TOKEN_SECRET"))
	refreshSecret = []byte(os.Getenv("REFRESH_TOKEN_SECRET"))
)

func main() {
	if len(accessSecret) == 0 || len(refreshSecret) == 0 {
		log.Fatalf("ACCESS_TOKEN_SECRET and REFRESH_TOKEN_SECRET must be set")
	}

	port := getEnv("GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	user_service_addr := getEnv("USER_SERVICE_URL", "user-service:50051")
	conn, err := grpc.NewClient(user_service_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to User Service: %v", err)
	}
	defer conn.Close()

	db_uri := getEnv("MARIADB_URI", "user:secret@tcp(auth-db:3306)/authdb")
	dsn := fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", db_uri)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Credentials{}, &RefreshToken{})

	authServer := &server{
		mariadbClient: db,
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, authServer)

	enable_reflection := getEnv("REFLECTION", "false")
	log.Println("Reflection enabled:", enable_reflection)
	if enable_reflection == "true" {
		reflection.Register(s)
	}

	log.Println("gRPC server started on port", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (s *server) RegisterCredentials(ctx context.Context, req *pb.RegisterCredentialsRequest) (*pb.RegisterCredentialsResponse, error) {
	hashedPassword, err := hashPassword(req.GetPassword())
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	user := Credentials{
		UserId:   req.GetUserId(),
		Username: req.GetUsername(),
		Password: hashedPassword,
	}

	if err := s.mariadbClient.Create(&user).Error; err != nil {
		log.Printf("Failed to register credentials: %v", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, status.Error(codes.AlreadyExists, "username already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register credentials")
	}

	return &pb.RegisterCredentialsResponse{}, nil
}

func (s *server) DeleteCredentials(ctx context.Context, req *pb.DeleteCredentialsRequest) (*pb.DeleteCredentialsResponse, error) {
	credToDelete := Credentials{UserId: req.GetUserId()}
	result := s.mariadbClient.WithContext(ctx).Delete(&credToDelete)

	if result.Error != nil {
		log.Printf("Failed to delete credentials: %v", result.Error)
		return nil, status.Error(codes.Internal, "failed to delete credentials")
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.DeleteCredentialsResponse{}, nil
}

func (s *server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	credentials := Credentials{Username: req.GetUsername()}
	if err := s.mariadbClient.Where(&credentials).First(&credentials).Error; err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid username")
	}
	if !checkPasswordHash(req.GetPassword(), credentials.Password) {
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	userId := credentials.UserId
	genericError := status.Error(codes.Internal, "failed to login")

	access, accessExp, err := signAccessToken(userId)
	if err != nil {
		log.Printf("failed to sign access token: %v", err)
		return nil, genericError
	}

	refresh, refreshExp, err := signRefreshToken(userId)
	if err != nil {
		log.Printf("failed to sign refresh token: %v", err)
		return nil, genericError
	}

	if err := saveRefreshToken(s.mariadbClient, userId, refresh, refreshExp); err != nil {
		log.Printf("failed to persist tokens: %v", err)
		return nil, genericError
	}

	return &pb.LoginResponse{
		Tokens: &pb.Tokens{
			AccessToken:           access,
			RefreshToken:          refresh,
			AccessTokenExpiresAt:  timestamppb.New(accessExp),
			RefreshTokenExpiresAt: timestamppb.New(refreshExp),
		},
	}, nil
}

func signAccessToken(userId string) (string, time.Time, error) {
	return issueToken(userId, accessTTL, accessSecret)
}

func signRefreshToken(userId string) (string, time.Time, error) {
	return issueToken(userId, refreshTTL, refreshSecret)
}

func issueToken(userId string, TTL time.Duration, secret []byte) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(TTL)
	claims := jwt.RegisteredClaims{
		Subject:   userId,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(secret)
	return s, exp, err
}

func saveRefreshToken(db *gorm.DB, userId, token string, exp time.Time) error {
	rt := &RefreshToken{
		UserId:    userId,
		Token:     token,
		ExpiresAt: exp,
	}
	return db.Save(&rt).Error
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) (err error) {
	if r.Id == "" {
		r.Id = uuid.New().String()
	}
	return
}

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
