package routes

import (
	"context"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
	stores *db.SQLiteStores
}

func NewAuthServer(stores *db.SQLiteStores) *AuthServer {
	return &AuthServer{stores: stores}
}

func (s *AuthServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "auth-routes").Logger()

	_, err := time.LoadLocation(req.GetTimezone())
	if err != nil {
		log.Warn().Str("timezone", req.GetTimezone()).Msg("invalid timezone")
		return nil, status.Errorf(codes.InvalidArgument, "invalid timezone: %v", req.GetTimezone())
	}

	id, err := s.stores.CreateUser(ctx, req.GetTimezone())
	if err != nil {
		log.Error().Err(err).Msg("failed to create user")
		return nil, err
	}
	log.Info().Str("user_id", id).Msg("user created")
	return &pb.SignUpResponse{
		UserId:    id,
		CreatedAt: time.Now().Unix(),
	}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "auth-routes").Logger()

	exists, err := s.stores.UserExists(ctx, req.GetUserId())
	if err != nil {
		log.Error().Err(err).Msg("failed to check user existence")
		return nil, err
	}
	log.Info().Str("user_id", req.GetUserId()).Bool("success", exists).Msg("login attempt")
	return &pb.LoginResponse{
		UserId:  req.GetUserId(),
		Success: exists,
	}, nil
}
