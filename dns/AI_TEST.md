!3 — Integration Tests (Valkey with testcontainers-go)
pkg/rcache/rcache_integration_test.go:
//go:build integration
func TestSetAndGet(t *testing.T) {
    ctx := context.Background()
    container, _ := testcontainers.GenericContainer(ctx, ...)  // valkey/valkey:9.0.3
    addr, _ := container.Endpoint(ctx, "")
    
    cache, _ := rcache.New(addr)
    err := cache.Set(ctx, cache.DomainDNSCacheKey("example.com"), responseBytes, 60)
    assert.NoError(t, err)
    // Once Get is added: verify the round-trip
}
Build tag integration keeps these out of standard go test ./... — run with go test -tags=integration ./....
---
Phase 4 — E2E Tests
Two tiers, both using build tags:
Standard E2E (fake upstream, controlled)  
//go:build e2e
- Start Valkey via testcontainers
- Start the full TLS server (RunServer) with a test cert + test context
- Connect a DNS client (using miekg/dns) to the server
- Server proxies to an in-process fake DoT server (a net.Listener that responds with canned DNS responses)
- Assert the client gets back a valid DNS response
Network E2E (real Cloudflare, requires internet)  
//go:build e2e_network
- Same as above but upstream is the real 1.1.1.1:853
- Run manually or in CI with internet access
- Use miekg/dns to query for google.com. A and assert a valid IP is returned
---
File Layout After All Changes
pkg/
  parser/
    parser.go
    parser_test.go            ← NEW: unit tests, miekg/dns fixtures on the fly
  proxy/
    proxy.go
    proxy_test.go             ← NEW: unit tests, in-process fake TLS server
  rcache/
    rcache.go                 ← MODIFIED: struct-based, injectable addr
    rcache_test.go            ← NEW: unit tests for key building, etc.
    rcache_integration_test.go ← NEW: //go:build integration, testcontainers
  server/
    server.go                 ← MODIFIED: DNSUpstream interface, context shutdown
    server_test.go            ← NEW: unit tests for handleDNSClient
    server_e2e_test.go        ← NEW: //go:build e2e
---
Dependencies to Add
go get github.com/testcontainers/testcontainers-go          # testcontainers
go get github.com/stretchr/testify                          # assert/require (optional but idiomatic)
No miniredis needed since you're using testcontainers.
---
Running Tests
# Unit tests only (fast, no Docker, no network)
go test ./...
# Integration tests (requires Docker)
go test -tags=integration ./...
# E2E (fake upstream, requires Docker for Valkey)
go test -tags=e2e ./...
# Full network E2E (requires Docker + internet)
go test -tags=e2e_network ./...
# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
---

