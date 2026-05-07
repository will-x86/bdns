package routes

import (
	"context"

	"codeberg.org/will-x86/bdns/dns/pkg/db"
	pb "codeberg.org/will-x86/bdns/dns/pkg/proto/pb"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileServer struct {
	pb.UnimplementedProfileServiceServer
	stores *db.SQLiteStores
}

func NewProfileServer(stores *db.SQLiteStores) *ProfileServer {
	return &ProfileServer{stores: stores}
}

func (s *ProfileServer) ListProfiles(ctx context.Context, req *pb.ListProfilesRequest) (*pb.ListProfilesResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "profile-routes").Logger()

	profiles, err := s.stores.ListProfiles(ctx, req.GetUserId())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("failed to list profiles")
		return nil, err
	}
	log.Info().Str("user_id", req.GetUserId()).Int("count", len(profiles)).Msg("profiles listed")

	var pbProfiles []*pb.Profile
	for _, p := range profiles {
		pbProfiles = append(pbProfiles, &pb.Profile{
			Id:        p.ID,
			UserId:    p.UserID,
			Name:      p.Name,
			CreatedAt: p.CreatedAt,
		})
	}
	return &pb.ListProfilesResponse{Profiles: pbProfiles}, nil
}

func (s *ProfileServer) CreateProfile(ctx context.Context, req *pb.CreateProfileRequest) (*pb.Profile, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "profile-routes").Logger()

	if req.GetName() == "" {
		log.Warn().Msg("profile name is empty")
		return nil, status.Errorf(codes.InvalidArgument, "profile name is required")
	}

	profileID, err := db.CreateProfile(req.GetUserId(), req.GetName())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Str("name", req.GetName()).Msg("failed to create profile")
		return nil, err
	}
	log.Info().Str("profile_id", profileID).Str("user_id", req.GetUserId()).Msg("profile created")

	profile, err := s.stores.GetProfile(ctx, profileID)
	if err != nil {
		log.Error().Err(err).Str("profile_id", profileID).Msg("failed to get created profile")
		return nil, err
	}
	return &pb.Profile{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt,
	}, nil
}

func (s *ProfileServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.Profile, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "profile-routes").Logger()

	profile, err := s.stores.GetProfile(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to get profile")
		return nil, err
	}
	if profile == nil {
		log.Warn().Str("profile_id", req.GetProfileId()).Msg("profile not found")
		return nil, status.Errorf(codes.NotFound, "profile not found")
	}
	log.Info().Str("profile_id", profile.ID).Msg("profile retrieved")
	return &pb.Profile{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt,
	}, nil
}

func (s *ProfileServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.Profile, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "profile-routes").Logger()

	if req.GetName() == "" {
		log.Warn().Msg("profile name is empty")
		return nil, status.Errorf(codes.InvalidArgument, "profile name is required")
	}

	profile, err := s.stores.UpdateProfile(ctx, req.GetProfileId(), req.GetName())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to update profile")
		return nil, err
	}
	if profile == nil {
		log.Warn().Str("profile_id", req.GetProfileId()).Msg("profile not found")
		return nil, status.Errorf(codes.NotFound, "profile not found")
	}
	log.Info().Str("profile_id", profile.ID).Msg("profile updated")
	return &pb.Profile{
		Id:        profile.ID,
		UserId:    profile.UserID,
		Name:      profile.Name,
		CreatedAt: profile.CreatedAt,
	}, nil
}

func (s *ProfileServer) DeleteProfile(ctx context.Context, req *pb.DeleteProfileRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "profile-routes").Logger()

	err := s.stores.DeleteProfile(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to delete profile")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Msg("profile deleted")
	return &emptypb.Empty{}, nil
}

