package rcache

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"codeberg.org/will-x86/bdns/dns/pkg/parser"
	"github.com/rs/zerolog"
	"github.com/valkey-io/valkey-go"
)

type DNSCache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl int64) error
	DomainDNSCacheKey(domain, qtype string) string
}

type Cache struct {
	client valkey.Client
}

func New(addr string) (*Cache, error) {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{addr}})
	if err != nil {
		return nil, err
	}
	return &Cache{client: client}, nil
}

func (c *Cache) DomainDNSCacheKey(domain, qtype string) string {
	return "dns_cache:" + domain + ":" + qtype
}

func (c *Cache) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	return c.client.Do(ctx, c.client.B().Set().Key(key).Value(string(value)).ExSeconds(ttl).Build()).Error()
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if resp.Error() != nil {
		if valkey.IsValkeyNil(resp.Error()) {
			return nil, nil // cache miss — not an error
		}
		return nil, resp.Error()
	}
	return resp.AsBytes()
}

// Returns the minimum TTL across all answer records in the response, or a default
func minTTLFromResponse(responseBytes []byte) int64 {
	const defaultTTL int64 = 60
	msg := parser.Message()
	if err := msg.Parse(responseBytes); err != nil {
		return defaultTTL
	}
	if len(msg.Answers) == 0 {
		return defaultTTL
	}
	min := int64(math.MaxInt64)
	for _, rr := range msg.Answers {
		if t := int64(rr.TTL); t < min {
			min = t
		}
	}
	if min <= 0 {
		return defaultTTL
	}
	return min
}

// Upstream + cache lookup, races them.
// Re-writes ID to have Transaction ID
// Cache on upstream win / nil
func (c *Cache) QueryRace(
	ctx context.Context,
	requestBytes []byte,
	domain, qtype string,
	upstream func(ctx context.Context) ([]byte, error),
) ([]byte, error) {
	return queryRace(ctx, c, requestBytes, domain, qtype, upstream)
}

func queryRace(
	ctx context.Context,
	store DNSCache,
	requestBytes []byte,
	domain, qtype string,
	upstream func(ctx context.Context) ([]byte, error),
) ([]byte, error) {
	log := zerolog.Ctx(ctx)
	if len(requestBytes) < 2 {
		return nil, fmt.Errorf("request too short to contain transaction ID")
	}
	reqID := binary.BigEndian.Uint16(requestBytes[0:2])

	type result struct {
		data   []byte
		err    error
		source string
	}

	ch := make(chan result, 2)

	// Cache lookup
	go func() {
		data, err := store.Get(ctx, store.DomainDNSCacheKey(domain, qtype))
		if err == nil && data == nil {
			// Normalise a cache miss so the upstream result wins automatically.
			err = errors.New("cache miss")
		}
		ch <- result{data, err, "cache"}
	}()

	// Upstream lookup
	go func() {
		data, err := upstream(ctx)
		ch <- result{data, err, "upstream"}
	}()

	var firstErr error
	for range 2 {
		r := <-ch
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
			}
			continue
		}
		if len(r.data) < 2 {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s returned a response too short to rewrite transaction ID", r.source)
			}
			continue
		}

		out := make([]byte, len(r.data))
		copy(out, r.data)

		// Rewrite transaction ID to match the client's request.
		binary.BigEndian.PutUint16(out[0:2], reqID)

		if r.source == "upstream" {
			log.Debug().Str("domain", domain).Str("q-type", qtype).Msg("upstream win")
			// Best effort
			_ = store.Set(ctx, store.DomainDNSCacheKey(domain, qtype), out, minTTLFromResponse(out))
		} else {
			log.Debug().Str("domain", domain).Str("q-type", qtype).Msg("cache win")
		}

		return out, nil
	}

	return nil, fmt.Errorf("both cache and upstream failed for %s %s: %w", domain, qtype, firstErr)
}
