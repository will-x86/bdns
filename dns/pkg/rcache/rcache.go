package rcache

import (
	"context"

	"github.com/valkey-io/valkey-go"
)

var client valkey.Client

func InitClient() error {
	var err error
	client, err = valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		return err
	}
	return nil

}
func DomainDNSCacheKey(domain string) string {
	return "dns_cache:" + domain
}

func Set(ctx context.Context, key string, value []byte, ttl int64) error {
	return client.Do(ctx, client.B().Set().Key(key).Value(string(value)).ExSeconds(ttl).Build()).Error()

}
