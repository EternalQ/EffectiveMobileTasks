package service

import (
	"context"
	"sync"

	_ "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/EternalQ/eff-mobile-tasks/task10/proto"
)

type GrpcService struct {
	pb.UnimplementedUserServiceServer

	mu     sync.RWMutex
	users  map[int32]*pb.User
	jwt    *JWTManager
	lastId int32
}

func NewGrpcService(jwt *JWTManager) *GrpcService {
	u := map[int32]*pb.User{0: {Id: 0, Name: "admin", Email: "admin"}}
	return &GrpcService{
		users:  u,
		lastId: 0,
		jwt:    jwt,
	}
}

func (s *GrpcService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "no creds")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var user *pb.User
	for _, v := range s.users {
		if v.Name == req.Name {
			user = v
		}
	}

	if user == nil {
		return nil, status.Error(codes.InvalidArgument, "wrong creds")
	}

	token, err := s.jwt.Generate(user.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "token generation error: %v", err)
	}

	return &pb.LoginResponse{Token: token}, nil
}

func (s *GrpcService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastId++
	user := &pb.User{
		Id:    s.lastId,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[user.Id] = user
	return user, nil
}

func (s *GrpcService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.Id]; !exists {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	user := &pb.User{
		Id:    req.Id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[req.Id] = user
	return user, nil
}

func (s *GrpcService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[req.Id]; !exists {
		return nil, status.Error(codes.NotFound, "User not found")
	}

	delete(s.users, req.Id)
	return &emptypb.Empty{}, nil
}

func (s *GrpcService) GetUsers(ctx context.Context, req *emptypb.Empty) (*pb.UserList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*pb.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return &pb.UserList{Users: users}, nil
}
