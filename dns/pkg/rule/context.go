package rule

import (
	"context"
	"time"

	"github.com/will-x86/bdns/dns/pkg/db/models"
)

type ProfileStore interface {
	ProfileExists(ctx context.Context, profileID string) (bool, error)
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

/*

type FriendshipStore interface {
	GetFriendship(ctx context.Context, userID string) (*models.Friendship, bool, error)
	DecrementAndCheck(ctx context.Context, friendshipID, userID string, poolSize int, date string) (bool, error)
}
*/

// CategoryResolver resolves a domain to a category (sits in rcache, injected here)
type CategoryResolver func(ctx context.Context, domain string) (string, error)

type Stores struct {
	Profile   ProfileStore
	Whitelist WhitelistStore
	Category  CategoryStore
	TimeBlock TimeBlockStore
	//	Friendship FriendshipStore
	Resolve CategoryResolver
}

// per-request - fields are lazily populated.
type RuleContext struct {
	Domain    string
	ProfileID string
	Now       time.Time

	// lazy — nil means not yet fetched
	category *string
	//friendship *models.Friendship

	Stores Stores
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
