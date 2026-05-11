package proto

import (
	"context"
	"net"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
	"codeberg.org/will-x86/bdns/dns/pkg/proto/routes"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func loggingInterceptor(log zerolog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		l := log.With().
			Str("method", info.FullMethod).
			Time("start", start).
			Logger()

		ctx = l.WithContext(ctx)

		resp, err := handler(ctx, req)

		code := codes.OK
		if err != nil {
			s, _ := status.FromError(err)
			code = s.Code()
			l.Error().Err(err).Str("code", code.String()).Dur("duration", time.Since(start)).Msg("request completed")
		} else {
			l.Info().Str("code", code.String()).Dur("duration", time.Since(start)).Msg("request completed")
		}

		return resp, err
	}
}

func RunServer(log zerolog.Logger) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(loggingInterceptor(log)))

	stores := db.NewStores(db.GetDB())
	pb.RegisterUserServiceServer(s, routes.NewUserServer(stores))
	pb.RegisterAuthServer(s, routes.NewAuthServer(stores))
	pb.RegisterProfileServiceServer(s, routes.NewProfileServer(stores))
	pb.RegisterWhitelistServiceServer(s, routes.NewWhitelistServer(stores))
	pb.RegisterCategoryServiceServer(s, routes.NewCategoryServer(stores))
	pb.RegisterTimeBlockServiceServer(s, routes.NewTimeBlockServer(stores))
	pb.RegisterPoolServiceServer(s, routes.NewPoolServer(stores))

	log.Info().Msg("gRPC server listening on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to serve")
	}
}
