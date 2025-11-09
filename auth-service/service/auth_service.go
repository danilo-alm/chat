package service

import (
	"auth-service/config"
	"auth-service/dto"
	"auth-service/repository"
	userpb "auth-service/user-pb"
	"context"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Login(ctx context.Context, username, rawPassword string) (*Tokens, error)
	RotateRefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)
}

type authService struct {
	repository  repository.AuthRepository
	userService userpb.UserServiceClient
	config      *config.Config
}

func NewAuthService(repository repository.AuthRepository, userService userpb.UserServiceClient, config *config.Config) AuthService {
	return &authService{
		repository:  repository,
		userService: userService,
		config:      config,
	}
}

func (s *authService) Login(ctx context.Context, username, rawPassword string) (*Tokens, error) {
	userReq := &userpb.GetCredentialsRequest{Username: username}
	userRes, err := s.userService.GetCredentials(ctx, userReq)
	if err != nil {
		return nil, err
	}

	if err = comparePassword(userRes.GetHashedPassword(), rawPassword); err != nil {
		return nil, err
	}

	claims := &claims{
		userId:   userRes.User.Id,
		username: userRes.User.Username,
		roles:    extractRoleNames(userRes.User.Roles),
	}
	tokens, err := s.generateTokens(claims)

	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *authService) RotateRefreshToken(ctx context.Context, oldToken string) (*Tokens, error) {
	decodedClaims, err := decodeToken(oldToken, s.config.RefreshSecret)
	if err != nil {
		return nil, err
	}

	if time.Now().After(decodedClaims.ExpiresAt.Time) {
		return nil, status.Error(codes.Unauthenticated, "refresh token expired")
	}

	claims := &claims{
		userId:   decodedClaims.Subject,
		username: decodedClaims.Username,
		roles:    decodedClaims.Roles,
	}

	newTokens, err := s.generateTokens(claims)
	if err != nil {
		return nil, err
	}

	rotateDto := &dto.SaveRefreshToken{
		RefreshToken: newTokens.Refresh,
		Expiration:   newTokens.RefreshExp,
	}
	err = s.repository.RotateRefreshToken(ctx, oldToken, rotateDto)

	if errors.Is(err, repository.ErrEntityNotFound) {
		return nil, status.Error(codes.Unauthenticated, "refresh token not found")
	} else if err != nil {
		log.Printf("failed to rotate refresh token: %v", err)
		return nil, status.Error(codes.Unauthenticated, "failed to refresh token")
	}

	return newTokens, nil
}

func (s *authService) generateTokens(c *claims) (*Tokens, error) {
	genericError := status.Error(codes.Internal, "failed to login")

	tc := &tokenConfig{
		accessSecret:  s.config.AccessSecret,
		refreshSecret: s.config.RefreshSecret,
		accessTTL:     s.config.AccessTTL,
		refreshTTL:    s.config.RefreshTTL,
	}

	access, accessExp, err := issueToken(c, tc.accessTTL, tc.accessSecret)
	if err != nil {
		log.Printf("failed to sign access token: %v", err)
		return nil, genericError
	}

	refresh, refreshExp, err := issueToken(c, tc.refreshTTL, tc.refreshSecret)
	if err != nil {
		log.Printf("failed to sign refresh token: %v", err)
		return nil, genericError
	}

	return &Tokens{
		Access:     access,
		AccessExp:  accessExp,
		Refresh:    refresh,
		RefreshExp: refreshExp,
	}, nil
}

func comparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return status.Error(codes.Unauthenticated, "wrong password")
	} else if err != nil {
		log.Printf("error hashing password: %v", err)
		return status.Error(codes.Internal, "could not login")
	}

	return nil
}

func extractRoleNames(roles []*userpb.Role) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, role.Name)
	}
	return names
}

func issueToken(c *claims, TTL time.Duration, secret []byte) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(TTL)
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   c.userId,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		Username: c.username,
		Roles:    c.roles,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(secret)
	return s, exp, err
}

func decodeToken(tokenString string, secret []byte) (*jwtClaims, error) {
	claims := &jwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "invalid signing method")
		}
		return secret, nil
	})

	if err != nil {
		log.Printf("failed to parse token: %v", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	if !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return claims, nil
}
