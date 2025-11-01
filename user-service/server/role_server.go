package server

import (
	"context"
	"user-service/dto"
	"user-service/pb"
	"user-service/service"
)

type RoleServer struct {
	pb.UnimplementedRoleServiceServer
	roleService service.RoleService
}

func NewRoleServer(roleService service.RoleService) *RoleServer {
	return &RoleServer{roleService: roleService}
}

func (s *RoleServer) CreateRole(ctx context.Context, req *pb.CreateRoleRequest) (*pb.CreateRoleResponse, error) {
	data := &dto.CreateRoleDto{
		Name: req.GetName(),
	}
	role, err := s.roleService.CreateRole(ctx, data)
	if err != nil {
		return nil, err
	}
	return &pb.CreateRoleResponse{Id: role.ID}, nil
}

func (s *RoleServer) CreateRoles(ctx context.Context, req *pb.CreateRolesRequest) (*pb.CreateRolesResponse, error) {
	roleDtos := make([]dto.CreateRoleDto, len(req.GetNames()))
	for i, name := range req.GetNames() {
		roleDtos[i].Name = name
	}
	roles, err := s.roleService.CreateRoles(ctx, roleDtos)
	if err != nil {
		return nil, err
	}

	rolesResponse := &pb.CreateRolesResponse{}
	rolesResponse.Ids = make([]string, len(roles))
	for i, role := range roles {
		rolesResponse.Ids[i] = role.ID
	}

	return rolesResponse, nil
}

func (s *RoleServer) GetRoleByName(ctx context.Context, req *pb.GetRoleByNameRequest) (*pb.GetRoleByNameResponse, error) {
	role, err := s.roleService.GetRoleByName(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &pb.GetRoleByNameResponse{
		Role: &pb.Role{
			Id:   role.ID,
			Name: role.Name,
		},
	}, nil
}

func (s *RoleServer) GetRolesByNames(ctx context.Context, req *pb.GetRolesByNamesRequest) (*pb.GetRolesByNamesResponse, error) {
	roles, err := s.roleService.GetRolesByNames(ctx, req.GetNames())
	if err != nil {
		return nil, err
	}

	rolesResponse := &pb.GetRolesByNamesResponse{}
	rolesResponse.Roles = make([]*pb.Role, len(roles))
	for i, role := range roles {
		rolesResponse.Roles[i] = &pb.Role{
			Id:   role.ID,
			Name: role.Name,
		}
	}

	return rolesResponse, nil
}

func (s *RoleServer) DeleteRoleById(ctx context.Context, req *pb.DeleteRoleByIdRequest) (*pb.DeleteRoleResponse, error) {
	err := s.roleService.DeleteRoleById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.DeleteRoleResponse{}, nil
}
