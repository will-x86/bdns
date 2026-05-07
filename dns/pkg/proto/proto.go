package proto

import (
	"log"
	"net"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
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

	stores := db.NewStores(db.GetDB())
	pb.RegisterUserServiceServer(s, routes.NewUserServer(stores))
	pb.RegisterAuthServer(s, routes.NewAuthServer(stores))

	log.Println("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
