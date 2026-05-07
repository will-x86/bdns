package routes

import (
	"context"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	stores *db.SQLiteStores
}

func NewUserServer(stores *db.SQLiteStores) *UserServer {
	return &UserServer{stores: stores}
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := s.stores.GetUser(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &pb.User{
		Id:        user.ID,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	user, err := s.stores.UpdateUser(ctx, req.GetUserId(), req.GetTimezone())
	if err != nil {
		return nil, err
	}
	return &pb.User{
		Id:        user.ID,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
	}, nil
}