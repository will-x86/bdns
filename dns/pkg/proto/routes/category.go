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

type CategoryServer struct {
	pb.UnimplementedCategoryServiceServer
	stores *db.SQLiteStores
}

func NewCategoryServer(stores *db.SQLiteStores) *CategoryServer {
	return &CategoryServer{stores: stores}
}

func (s *CategoryServer) ListBlocked(ctx context.Context, req *pb.ListBlockedRequest) (*pb.CategoryList, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "category-routes").Logger()

	categories, err := s.stores.ListBlockedCategories(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to list blocked categories")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Int("count", len(categories)).Msg("listed blocked categories")
	return &pb.CategoryList{Categories: categories}, nil
}

func (s *CategoryServer) BlockCategory(ctx context.Context, req *pb.BlockCategoryRequest) (*pb.CategoryBlock, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "category-routes").Logger()

	if req.GetCategory() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "category is required")
	}
	err := s.stores.BlockCategory(ctx, req.GetProfileId(), req.GetCategory())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("category", req.GetCategory()).Msg("failed to block category")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("category", req.GetCategory()).Msg("blocked category")
	return &pb.CategoryBlock{Category: req.GetCategory()}, nil
}

func (s *CategoryServer) UnblockCategory(ctx context.Context, req *pb.UnblockCategoryRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "category-routes").Logger()

	err := s.stores.UnblockCategory(ctx, req.GetProfileId(), req.GetCategory())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("category", req.GetCategory()).Msg("failed to unblock category")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("category", req.GetCategory()).Msg("unblocked category")
	return &emptypb.Empty{}, nil
}