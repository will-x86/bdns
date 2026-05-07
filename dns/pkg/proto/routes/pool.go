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

type PoolServer struct {
	pb.UnimplementedPoolServiceServer
	stores *db.SQLiteStores
}

func NewPoolServer(stores *db.SQLiteStores) *PoolServer {
	return &PoolServer{stores: stores}
}

func (s *PoolServer) ListPools(ctx context.Context, req *pb.ListPoolsRequest) (*pb.PoolList, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	pools, err := s.stores.ListPoolsForUser(ctx, req.GetUserId())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("failed to list pools")
		return nil, err
	}
	log.Info().Str("user_id", req.GetUserId()).Int("count", len(pools)).Msg("listed pools")

	var pbPools []*pb.FriendPool
	for _, p := range pools {
		pbPools = append(pbPools, &pb.FriendPool{
			Id:         p.ID,
			CreatedBy:  p.CreatedBy,
			Name:       p.Name,
			PoolMode:   p.PoolMode,
			TotalLimit: p.TotalLimit,
		})
	}
	return &pb.PoolList{Pools: pbPools}, nil
}

func (s *PoolServer) CreatePool(ctx context.Context, req *pb.CreatePoolRequest) (*pb.FriendPool, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	if req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	if req.GetPoolMode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "pool_mode is required")
	}
	id, err := s.stores.CreatePool(ctx, req.GetUserId(), req.GetName(), req.GetPoolMode(), req.GetTotalLimit())
	if err != nil {
		log.Error().Err(err).Str("user_id", req.GetUserId()).Msg("failed to create pool")
		return nil, err
	}
	log.Info().Str("pool_id", id).Msg("created pool")
	return &pb.FriendPool{
		Id:         id,
		CreatedBy:  req.GetUserId(),
		Name:       req.GetName(),
		PoolMode:   req.GetPoolMode(),
		TotalLimit: req.GetTotalLimit(),
	}, nil
}

func (s *PoolServer) GetPool(ctx context.Context, req *pb.GetPoolRequest) (*pb.FriendPool, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	pool, err := s.stores.GetPool(ctx, req.GetPoolId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Msg("failed to get pool")
		return nil, err
	}
	log.Info().Str("pool_id", pool.ID).Msg("got pool")
	return &pb.FriendPool{
		Id:         pool.ID,
		CreatedBy:  pool.CreatedBy,
		Name:       pool.Name,
		PoolMode:   pool.PoolMode,
		TotalLimit: pool.TotalLimit,
	}, nil
}

func (s *PoolServer) DeletePool(ctx context.Context, req *pb.DeletePoolRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	err := s.stores.DeletePool(ctx, req.GetPoolId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Msg("failed to delete pool")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Msg("deleted pool")
	return &emptypb.Empty{}, nil
}

func (s *PoolServer) JoinPool(ctx context.Context, req *pb.JoinPoolRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	err := s.stores.JoinPool(ctx, req.GetPoolId(), req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Str("profile_id", req.GetProfileId()).Msg("failed to join pool")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Str("profile_id", req.GetProfileId()).Msg("joined pool")
	return &emptypb.Empty{}, nil
}

func (s *PoolServer) LeavePool(ctx context.Context, req *pb.LeavePoolRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	err := s.stores.LeavePool(ctx, req.GetPoolId(), req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Str("profile_id", req.GetProfileId()).Msg("failed to leave pool")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Str("profile_id", req.GetProfileId()).Msg("left pool")
	return &emptypb.Empty{}, nil
}

func (s *PoolServer) ListMembers(ctx context.Context, req *pb.ListMembersRequest) (*pb.PoolMemberList, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	members, err := s.stores.ListPoolMembers(ctx, req.GetPoolId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Msg("failed to list pool members")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Int("count", len(members)).Msg("listed pool members")

	var pbMembers []*pb.PoolMember
	for _, m := range members {
		pbMembers = append(pbMembers, &pb.PoolMember{
			PoolId:    m.PoolID,
			ProfileId: m.ProfileID,
		})
	}
	return &pb.PoolMemberList{Members: pbMembers}, nil
}

func (s *PoolServer) ListBlocks(ctx context.Context, req *pb.ListPoolBlocksRequest) (*pb.PoolBlockList, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	categories, err := s.stores.ListPoolCategoryBlocks(ctx, req.GetPoolId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Msg("failed to list pool category blocks")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Int("count", len(categories)).Msg("listed pool category blocks")

	var pbBlocks []*pb.PoolBlock
	for _, c := range categories {
		pbBlocks = append(pbBlocks, &pb.PoolBlock{
			Category: c,
		})
	}
	return &pb.PoolBlockList{Blocks: pbBlocks}, nil
}

func (s *PoolServer) BlockCategory(ctx context.Context, req *pb.BlockPoolCategoryRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	err := s.stores.AddPoolCategoryBlock(ctx, req.GetPoolId(), req.GetCategory())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Str("category", req.GetCategory()).Msg("failed to block pool category")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Str("category", req.GetCategory()).Msg("blocked pool category")
	return &emptypb.Empty{}, nil
}

func (s *PoolServer) UnblockCategory(ctx context.Context, req *pb.UnblockPoolCategoryRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	err := s.stores.RemovePoolCategoryBlock(ctx, req.GetPoolId(), req.GetCategory())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Str("category", req.GetCategory()).Msg("failed to unblock pool category")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Str("category", req.GetCategory()).Msg("unblocked pool category")
	return &emptypb.Empty{}, nil
}

func (s *PoolServer) GetCredits(ctx context.Context, req *pb.GetCreditsRequest) (*pb.CreditsResponse, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "pool-routes").Logger()

	credits, err := s.stores.GetPoolCredits(ctx, req.GetPoolId())
	if err != nil {
		log.Error().Err(err).Str("pool_id", req.GetPoolId()).Msg("failed to get pool credits")
		return nil, err
	}
	log.Info().Str("pool_id", req.GetPoolId()).Int64("credits", credits).Msg("got pool credits")
	return &pb.CreditsResponse{Remaining: credits, Total: credits}, nil
}