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

type WhitelistServer struct {
	pb.UnimplementedWhitelistServiceServer
	stores *db.SQLiteStores
}

func NewWhitelistServer(stores *db.SQLiteStores) *WhitelistServer {
	return &WhitelistServer{stores: stores}
}

func (s *WhitelistServer) ListPermanent(ctx context.Context, req *pb.ListPermanentRequest) (*pb.WhitelistDomains, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	domains, err := s.stores.ListPermanentWhitelists(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to list permanent whitelists")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Int("count", len(domains)).Msg("listed permanent whitelists")
	return &pb.WhitelistDomains{Domains: domains}, nil
}

func (s *WhitelistServer) AddPermanent(ctx context.Context, req *pb.AddPermanentRequest) (*pb.WhitelistDomain, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	if req.GetDomain() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "domain is required")
	}
	err := s.stores.AddPermanentWhitelist(ctx, req.GetProfileId(), req.GetDomain())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("failed to add permanent whitelist")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("added permanent whitelist")
	return &pb.WhitelistDomain{Domain: req.GetDomain()}, nil
}

func (s *WhitelistServer) RemovePermanent(ctx context.Context, req *pb.RemovePermanentRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	err := s.stores.RemovePermanentWhitelist(ctx, req.GetProfileId(), req.GetDomain())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("failed to remove permanent whitelist")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("removed permanent whitelist")
	return &emptypb.Empty{}, nil
}

func (s *WhitelistServer) ListTemporary(ctx context.Context, req *pb.ListTemporaryRequest) (*pb.WhitelistDomainsTemp, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	whitelists, err := s.stores.ListTemporaryWhitelists(ctx, req.GetProfileId())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Msg("failed to list temporary whitelists")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Int("count", len(whitelists)).Msg("listed temporary whitelists")

	var entries []*pb.WhitelistDomainTemp
	for _, w := range whitelists {
		entries = append(entries, &pb.WhitelistDomainTemp{
			Domain:    w.Domain,
			ExpiresAt: int64(w.ExpiresAt),
		})
	}
	return &pb.WhitelistDomainsTemp{Entries: entries}, nil
}

func (s *WhitelistServer) AddTemporary(ctx context.Context, req *pb.AddTemporaryRequest) (*pb.WhitelistDomainTemp, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	if req.GetDomain() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "domain is required")
	}
	if req.GetExpiresAt() == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "expires_at is required")
	}
	err := s.stores.AddTemporaryWhitelist(ctx, req.GetProfileId(), req.GetDomain(), req.GetExpiresAt())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("failed to add temporary whitelist")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Int64("expires_at", req.GetExpiresAt()).Msg("added temporary whitelist")
	return &pb.WhitelistDomainTemp{
		Domain:    req.GetDomain(),
		ExpiresAt: req.GetExpiresAt(),
	}, nil
}

func (s *WhitelistServer) RemoveTemporary(ctx context.Context, req *pb.RemoveTemporaryRequest) (*emptypb.Empty, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "whitelist-routes").Logger()

	err := s.stores.RemoveTemporaryWhitelist(ctx, req.GetProfileId(), req.GetDomain())
	if err != nil {
		log.Error().Err(err).Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("failed to remove temporary whitelist")
		return nil, err
	}
	log.Info().Str("profile_id", req.GetProfileId()).Str("domain", req.GetDomain()).Msg("removed temporary whitelist")
	return &emptypb.Empty{}, nil
}