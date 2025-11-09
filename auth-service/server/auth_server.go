package server

import (
	"auth-service/pb"
	"auth-service/service"
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthServer(authService service.AuthService) *AuthServer {
	return &AuthServer{
		authService: authService,
	}
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	tokens, err := s.authService.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Tokens: tokensToProtoTokens(tokens),
	}, nil
}

func (s *AuthServer) RotateRefreshToken(ctx context.Context, req *pb.RotateRefreshTokenRequest) (*pb.RotateRefreshTokenResponse, error) {
	tokens, err := s.authService.RotateRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &pb.RotateRefreshTokenResponse{
		Tokens: tokensToProtoTokens(tokens),
	}, nil
}

func tokensToProtoTokens(t *service.Tokens) *pb.Tokens {
	return &pb.Tokens{
		AccessToken:           t.Access,
		RefreshToken:          t.Refresh,
		AccessTokenExpiresAt:  timestamppb.New(t.AccessExp),
		RefreshTokenExpiresAt: timestamppb.New(t.RefreshExp),
	}
}
