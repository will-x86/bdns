package rcache

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

type Cache struct {
	client valkey.Client
}

func New(addr string) (*Cache, error) {

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return &Cache{}, err
	}

	return &Cache{
		client: client,
	}, nil

}
func (c *Cache) DomainDNSCacheKey(domain string) string {
	return "dns_cache:" + domain
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(string(value)).ExSeconds(ttl).Build()).Error()

}
