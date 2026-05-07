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

type TimeBlockServer struct {
	pb.UnimplementedTimeBlockServiceServer
	stores *db.SQLiteStores
}

func NewTimeBlockServer(stores *db.SQLiteStores) *TimeBlockServer {
	return &TimeBlockServer{stores: stores}
}

func (s *TimeBlockServer) List(ctx context.Context, req *pb.ListTimeBlocksRequest) (*pb.TimeBlockList, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "timeblock-routes").Logger()

	timeblocks, err := s.stores.ListTimeBlocks(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to list time blocks")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Int("count", len(timeblocks)).Msg("listed time blocks")

	var pbBlocks []*pb.TimeBlock
	for _, tb := range timeblocks {
		pbBlocks = append(pbBlocks, &pb.TimeBlock{
			ProfileId: tb.ProfileID,
			Category:  tb.Category,
			StartTime: int32(tb.StartTime),
			EndTime:   int32(tb.EndTime),
			Day:       int32(tb.Day),
			CreatedAt: int64(tb.CreatedAt),
		})
	}
	return &pb.TimeBlockList{Blocks: pbBlocks}, nil
}

func (s *TimeBlockServer) Create(ctx context.Context, req *pb.CreateTimeBlockRequest) (*pb.TimeBlock, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "timeblock-routes").Logger()

	if req.GetCategory() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "category is required")
	}
	err := s.stores.CreateTimeBlock(ctx, req.GetProfileId(), req.GetCategory(), int(req.GetStartTime()), int(req.GetEndTime()), int(req.GetDay()))
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to create time block")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("category", req.GetCategory()).Msg("created time block")
	return &pb.TimeBlock{
		ProfileId: req.GetProfileId(),
		Category:  req.GetCategory(),
		StartTime: req.GetStartTime(),
		EndTime:   req.GetEndTime(),
		Day:       req.GetDay(),
	}, nil
}

func (s *TimeBlockServer) Delete(ctx context.Context, req *pb.DeleteTimeBlockRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "timeblock-routes").Logger()

	err := s.stores.DeleteTimeBlock(ctx, req.GetBlockId(), "", 0, 0, 0)
	if err != nil {
		log.Error().Err(err).Str("block_id", req.GetBlockId()).Msg("failed to delete time block")
		return nil, err
	}
	log.Info().Str("block_id", req.GetBlockId()).Msg("deleted time block")
	return &emptypb.Empty{}, nil
}