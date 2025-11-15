package service

import (
	"auth-service/config"
	"auth-service/dto"
	"auth-service/repository"
	userpb "auth-service/user-pb"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	pbRes, err := s.userService.GetCredentials(ctx, userReq)
	if err != nil {
		return nil, err
	}

	if err = comparePassword(pbRes.GetHashedPassword(), rawPassword); err != nil {
		return nil, err
	}

	user := pbRes.GetUser()
	claims := &claims{
		userId:   user.Id,
		username: user.Username,
		roles:    extractRoleNames(user.Roles),
	}

	tokens, err := s.generateTokens(claims)
	if err != nil {
		return nil, err
	}

	if err = s.saveRefreshToken(ctx, tokens.Refresh, user.Id, tokens.RefreshExp); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *authService) RotateRefreshToken(ctx context.Context, oldToken string) (*Tokens, error) {
	token, err := s.repository.GetRefreshToken(ctx, oldToken)
	if errors.Is(err, repository.ErrEntityNotFound) {
		return nil, status.Error(codes.Unauthenticated, "refresh token not found")
	} else if err != nil {
		log.Printf("failed to get refresh token: %v", err)
		return nil, status.Error(codes.Unauthenticated, "failed to refresh token")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, status.Error(codes.Unauthenticated, "refresh token expired")
	}

	userReq := &userpb.GetUserByIdRequest{Id: token.UserID}
	userRes, err := s.userService.GetUserById(ctx, userReq)
	if err != nil {
		log.Printf("failed to get user: %v", err)
		return nil, status.Error(codes.Unauthenticated, "failed to refresh token")
	}

	claims := &claims{
		userId:   token.UserID,
		username: userRes.User.Username,
		roles:    extractRoleNames(userRes.User.Roles),
	}
	newTokens, err := s.generateTokens(claims)
	if err != nil {
		return nil, err
	}

	rotateDto := &dto.SaveRefreshToken{
		RefreshToken: newTokens.Refresh,
		UserID:       token.UserID,
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
	cfg := *s.config

	access, accessExp, err := issueJwtToken(c, cfg.AccessTTL, cfg.AccessSecret)
	if err != nil {
		log.Printf("failed to sign access token: %v", err)
		return nil, status.Error(codes.Internal, "failed to login")
	}

	refresh, refreshExp := issueOpaqueToken(cfg.RefreshTTL)

	return &Tokens{
		Access:     access,
		AccessExp:  accessExp,
		Refresh:    refresh,
		RefreshExp: refreshExp,
	}, nil
}

func (s *authService) saveRefreshToken(ctx context.Context, token, userID string, expiration time.Time) error {
	saveDto := &dto.SaveRefreshToken{
		RefreshToken: token,
		UserID:       userID,
		Expiration:   expiration,
	}

	if err := s.repository.SaveRefreshToken(ctx, saveDto); err != nil {
		log.Printf("failed to save refresh token: %v", err)
		return status.Error(codes.Internal, "could not issue refresh token")
	}

	return nil
}

func issueOpaqueToken(TTL time.Duration) (string, time.Time) {
	token := strings.ReplaceAll(uuid.NewString(), "-", "")
	expiration := time.Now().Add(TTL)
	return token, expiration
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

func issueJwtToken(c *claims, TTL time.Duration, secret []byte) (string, time.Time, error) {
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
