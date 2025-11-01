package server

import (
	"context"
	"user-service/dto"
	"user-service/models"
	"user-service/pb"
	"user-service/service"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	userService service.UserService
}

func NewUserServer(userService service.UserService, roleService service.RoleService) *UserServer {
	return &UserServer{userService: userService}
}

func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	data := &dto.RegisterUserDto{
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	}
	userId, err := s.userService.RegisterUser(ctx, data)
	if err != nil {
		return nil, err
	}
	return &pb.CreateUserResponse{Id: userId}, nil
}

func (s *UserServer) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserResponse, error) {
	user, err := s.userService.GetUserById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return mapUserToPbUserResponse(user), nil
}

func (s *UserServer) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.GetUserResponse, error) {
	user, err := s.userService.GetUserByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, err
	}
	return mapUserToPbUserResponse(user), nil
}

func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
	if err := s.userService.DeleteUserById(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteUserResponse{}, nil
}

func (s *UserServer) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	if err := s.userService.AssignRole(ctx, req.GetUserId(), req.GetRoleName()); err != nil {
		return nil, err
	}
	return &pb.AssignRoleResponse{}, nil
}

func mapUserToPbUserResponse(user *models.User) *pb.GetUserResponse {
	return &pb.GetUserResponse{User: &pb.User{
		Id:       user.ID,
		Name:     user.Name,
		Username: user.Username,
	}}
}
