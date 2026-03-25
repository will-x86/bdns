package rule

import (
	"context"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/db/models"
)

type ProfileStore interface {
	ProfileExists(ctx context.Context, profileID string) (bool, error)
	GetProfileWithUser(ctx context.Context, profileID string) (*models.Profile, *models.User, error)
}

type WhitelistStore interface {
	IsPermanentlyWhitelisted(ctx context.Context, profileID, domain string) (bool, error)
	IsTemporarilyWhitelisted(ctx context.Context, profileID, domain string, now time.Time) (bool, error)
}

type CategoryStore interface {
	IsCategoryBlocked(ctx context.Context, profileID, category string) (bool, error)
}

type TimeBlockStore interface {
	GetTimeBlocks(ctx context.Context, profileID, category string) ([]models.TimeBlock, error)
}

type PoolCacheStore interface {
	PoolID(ctx context.Context, profileID string) (string, error)
	ExistsShared(ctx context.Context, poolID string) bool
	ExistsBorrow(ctx context.Context, poolID, profileID string) bool
	DecrementRemainingBorrow(ctx context.Context, poolID, profileID string) error
	GetRemainingShared(ctx context.Context, poolID string) (int64, error)
	DecrementRemainingShared(ctx context.Context, poolID string) error
	GetRemainingBorrow(ctx context.Context, poolID, profileID string) (int64, error)
}

type PoolDBStore interface {
	GetPool(ctx context.Context, poolID string) (models.FriendPool, error)
	PoolCategoryBlocked(ctx context.Context, poolID, category string) bool
}

// CategoryResolver resolves a domain to a category (sits in rcache, injected here)
type CategoryResolver func(ctx context.Context, domain string) (string, error)

type Stores struct {
	Profile   ProfileStore
	Whitelist WhitelistStore
	Category  CategoryStore
	TimeBlock TimeBlockStore
	PoolCache PoolCacheStore
	PoolDB    PoolDBStore
	Resolve   CategoryResolver
}

type RuleContext struct {
	Domain    string
	ProfileID string
	Now       time.Time
	Profile   *models.Profile
	User      *models.User

	category *string
	pools    map[string]*models.FriendPool

	Stores Stores
}

func (r *RuleContext) GetPool(ctx context.Context, poolID string) (models.FriendPool, error) {
	if r.pools == nil {
		r.pools = make(map[string]*models.FriendPool)
	}
	if p, ok := r.pools[poolID]; ok {
		return *p, nil
	}
	pol, err := r.Stores.PoolDB.GetPool(ctx, poolID)
	if err != nil {
		return models.FriendPool{}, err
	}
	r.pools[poolID] = &pol
	return pol, nil
}
func (r *RuleContext) GetCategory(ctx context.Context) (string, error) {
	if r.category != nil {
		return *r.category, nil
	}
	cat, err := r.Stores.Resolve(ctx, r.Domain)
	if err != nil {
		return "", err
	}
	r.category = &cat
	return cat, nil
}

/*func (r *RuleContext) GetFriendship(ctx context.Context) (*models.Friendship, bool, error) {
	if r.friendship != nil {
		return r.friendship, true, nil
	}
	f, ok, err := r.Stores.Friendship.GetFriendship(ctx, r.UserID)
	if err != nil || !ok {
		return nil, false, err
	}
	r.friendship = f
	return f, true, nil
}
*/
