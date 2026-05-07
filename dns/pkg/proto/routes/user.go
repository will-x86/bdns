package routes

import (
	"context"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
	"github.com/rs/zerolog"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	stores *db.SQLiteStores
}

func NewUserServer(stores *db.SQLiteStores) *UserServer {
	return &UserServer{stores: stores}
}

func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "user-routes").Logger()

	user, err := s.stores.GetUser(ctx, req.GetUserId())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("failed to get user")
		return nil, err
	}
	log.Info().Str("user_id", user.ID).Msg("user retrieved")
	return &pb.User{
		Id:        user.ID,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "user-routes").Logger()

	user, err := s.stores.UpdateUser(ctx, req.GetUserId(), req.GetTimezone())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("failed to update user")
		return nil, err
	}
	log.Info().Str("user_id", user.ID).Msg("user updated")
	return &pb.User{
		Id:        user.ID,
		Timezone:  user.Timezone,
		CreatedAt: user.CreatedAt,
	}, nil
}