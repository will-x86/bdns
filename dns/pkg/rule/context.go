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
	PoolID(ctx context.Context, profileID string) (string, error) // Returns pool_id with profile_id
	Exists(ctx context.Context, profileID, poolID string) bool    // Sees if a pool exists with
	// Decrement // increment
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
	pool     *models.FriendPool

	Stores Stores
}

func (r *RuleContext) GetPool(ctx context.Context, poolID string) (models.FriendPool, error) {
	if r.pool != nil {
		return *r.pool, nil
	}
	pol, err := r.Stores.PoolDB.GetPool(ctx, poolID)
	if err != nil {
		return models.FriendPool{}, err
	}
	r.pool = &pol
	return *r.pool, nil
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
