package routes

import (
	"context"
	"log"

	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log.Printf("user ID %s", req.GetUserId())
	return &pb.User{
		Id:       "ID I guess",
		Timezone: "idk yet",
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	log.Printf("user ID %s", req.GetUserId())
	return &pb.User{
		Id:       "ID changed I guess",
		Timezone: "idk yet",
	}, nil
}
