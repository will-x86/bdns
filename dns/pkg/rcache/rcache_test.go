package rcache

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/valkey"
)

// test queryRace without real valkey
type fakeStore struct {
	mu   sync.Mutex
	data map[string][]byte
	// getErr, if set, is returned by Get instead of looking up data.
	getErr error
	// setErr, if set, is returned by Set.
	setErr error
	// setCalls records every (key, value, ttl) passed to Set.
	setCalls []setCall
}

type setCall struct {
	key   string
	value []byte
	ttl   int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{data: make(map[string][]byte)}
}

func (f *fakeStore) DomainDNSCacheKey(domain, qtype string) string {
	return "dns_cache:" + domain + ":" + qtype
}

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	v, ok := f.data[key]
	if !ok {
		return nil, nil // cache miss
	}
	out := make([]byte, len(v))
	copy(out, v)
	return out, nil
}

func (f *fakeStore) Set(_ context.Context, key string, value []byte, ttl int64) error {
	if f.setErr != nil {
		return f.setErr
	}
	cp := make([]byte, len(value))
	copy(cp, value)
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = cp
	f.setCalls = append(f.setCalls, setCall{key, cp, ttl})
	return nil
}

// return 12-byte response header with given ID & no records
func minimalResponse(id uint16) []byte {
	b := make([]byte, 12)
	binary.BigEndian.PutUint16(b[0:2], id)
	binary.BigEndian.PutUint16(b[2:4], 0x8180) // QR=1 RD=1 RA=1
	return b
}

// extract transaction ID
func txID(b []byte) uint16 {
	return binary.BigEndian.Uint16(b[0:2])
}

func TestDomainDNSCacheKey(t *testing.T) {
	c := &Cache{}
	got := c.DomainDNSCacheKey("example.com", "A")
	want := "dns_cache:example.com:A"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Should feault to 60
func TestMinTTLFromResponse_NoAnswers(t *testing.T) {
	if got := minTTLFromResponse(minimalResponse(1)); got != 60 {
		t.Errorf("expected default 60, got %d", got)
	}
}

func TestMinTTLFromResponse_TooShort(t *testing.T) {
	if got := minTTLFromResponse([]byte{0x00}); got != 60 {
		t.Errorf("expected default 60 for unparseable input, got %d", got)
	}
}

// Query race

func TestQueryRace_RequestTooShort(t *testing.T) {
	_, err := queryRace(context.Background(), newFakeStore(), []byte{0x01}, "a.com", "A",
		func(_ context.Context) ([]byte, error) { return minimalResponse(1), nil },
	)
	if err == nil {
		t.Fatal("expected error for request shorter than 2 bytes")
	}
}

// Cache misses, upstream result is returned & replaced
func TestQueryRace_UpstreamWins_IDRewritten(t *testing.T) {
	const reqID uint16 = 0xABCD
	const upstreamID uint16 = 0x1234

	got, err := queryRace(context.Background(), newFakeStore(), minimalResponse(reqID), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return minimalResponse(upstreamID), nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txID(got) != reqID {
		t.Errorf("response ID = %#x, want %#x (request ID)", txID(got), reqID)
	}
}

func TestQueryRace_UpstreamWins_StoredInCache(t *testing.T) {
	store := newFakeStore()
	_, err := queryRace(context.Background(), store, minimalResponse(0x0001), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return minimalResponse(0x9999), nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	store.mu.Lock()
	calls := store.setCalls
	store.mu.Unlock()

	if len(calls) != 1 {
		t.Fatalf("expected 1 Set call, got %d", len(calls))
	}
	if calls[0].ttl <= 0 {
		t.Errorf("expected positive TTL, got %d", calls[0].ttl)
	}
	if want := store.DomainDNSCacheKey("example.com", "A"); calls[0].key != want {
		t.Errorf("stored under key %q, want %q", calls[0].key, want)
	}
}

// cached response has a stale ID; the
// returned bytes must carry the current request's ID.
func TestQueryRace_CacheHit_IDRewritten(t *testing.T) {
	const reqID uint16 = 0xBEEF
	const cachedID uint16 = 0x0001

	store := newFakeStore()
	key := store.DomainDNSCacheKey("example.com", "A")
	_ = store.Set(context.Background(), key, minimalResponse(cachedID), 60)
	store.mu.Lock()
	store.setCalls = nil
	store.mu.Unlock()

	// upstream always fails so we know the cache result must be returned
	got, err := queryRace(context.Background(), store, minimalResponse(reqID), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return nil, errors.New("upstream down") },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txID(got) != reqID {
		t.Errorf("response ID = %#x, want %#x (request ID)", txID(got), reqID)
	}
}

// a cache hit must not trigger a Set call.
func TestQueryRace_CacheHit_NoExtraSet(t *testing.T) {
	store := newFakeStore()
	key := store.DomainDNSCacheKey("example.com", "A")
	_ = store.Set(context.Background(), key, minimalResponse(0x0001), 60)
	store.mu.Lock()
	store.setCalls = nil
	store.mu.Unlock()

	_, err := queryRace(context.Background(), store, minimalResponse(0x0042), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return nil, errors.New("upstream down") },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	store.mu.Lock()
	calls := store.setCalls
	store.mu.Unlock()

	if len(calls) != 0 {
		t.Errorf("expected no Set calls on cache hit, got %d", len(calls))
	}
}

// error returned when both cache and upstream fail
func TestQueryRace_BothFail(t *testing.T) {
	store := newFakeStore()
	store.getErr = errors.New("valkey unreachable")

	_, err := queryRace(context.Background(), store, minimalResponse(0x0001), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return nil, errors.New("upstream timeout") },
	)
	if err == nil {
		t.Fatal("expected error when both cache and upstream fail")
	}
}

// a cache write failure must not
// prevent the upstream response from being returned to the caller.
func TestQueryRace_SetFailureDoesNotBlockResponse(t *testing.T) {
	const reqID uint16 = 0x0007
	store := newFakeStore()
	store.setErr = errors.New("disk full")

	got, err := queryRace(context.Background(), store, minimalResponse(reqID), "example.com", "A",
		func(_ context.Context) ([]byte, error) { return minimalResponse(0x1111), nil },
	)
	if err != nil {
		t.Fatalf("Set failure should not cause an error, got: %v", err)
	}
	if txID(got) != reqID {
		t.Errorf("response ID = %#x, want %#x", txID(got), reqID)
	}
}

// queryRace must not modify the
// caller's request slice or the upstream's response slice.
func TestQueryRace_OriginalBytesNotMutated(t *testing.T) {
	request := minimalResponse(0xAAAA)
	requestSnap := make([]byte, len(request))
	copy(requestSnap, request)

	upstreamResp := minimalResponse(0xBBBB)
	upstreamSnap := make([]byte, len(upstreamResp))
	copy(upstreamSnap, upstreamResp)

	_, err := queryRace(context.Background(), newFakeStore(), request, "example.com", "A",
		func(_ context.Context) ([]byte, error) { return upstreamResp, nil },
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := range requestSnap {
		if request[i] != requestSnap[i] {
			t.Errorf("requestBytes mutated at [%d]: got %#x want %#x", i, request[i], requestSnap[i])
		}
	}
	for i := range upstreamSnap {
		if upstreamResp[i] != upstreamSnap[i] {
			t.Errorf("upstream response mutated at [%d]: got %#x want %#x", i, upstreamResp[i], upstreamSnap[i])
		}
	}
}

// Integration tests
func setupValkeyContainer(ctx context.Context) (*valkey.ValkeyContainer, error) {
	return valkey.Run(ctx, "docker.io/valkey/valkey:9.0.3")
}

func TestRCache(t *testing.T) {
	ctx := context.Background()
	valkeyContainer, err := setupValkeyContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to start Valkey container: %v", err)
	}
	defer valkeyContainer.Terminate(ctx)

	host, err := valkeyContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Valkey host: %v", err)
	}
	port, err := valkeyContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Valkey mapped port: %v", err)
	}

	cache, err := New(host + ":" + port.Port())
	if err != nil {
		t.Fatalf("Failed to create Cache: %v", err)
	}

	key := cache.DomainDNSCacheKey("example.com", "A")
	value := []byte("testValue")

	if err := cache.Set(ctx, key, value, 100); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(got) != string(value) {
		t.Errorf("Get returned %q, want %q", got, value)
	}

	// Miss returns nil, nil
	missing, err := cache.Get(ctx, "dns_cache:missing.com:A")
	if err != nil {
		t.Fatalf("Get for missing key returned error: %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for missing key, got %v", missing)
	}
}
