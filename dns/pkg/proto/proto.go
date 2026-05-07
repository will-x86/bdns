package proto

import (
	"log"
	"net"

	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
	"codeberg.org/will-x86/bdns/dns/pkg/proto/routes"
	"google.golang.org/grpc"
)

func RunServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &routes.UserServer{})

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
