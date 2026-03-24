Part 1:
Set TTL for 4am in timezone 
When expires, it's gone
Part 2:
Add KEEPTTL to all sets / decrs
Part 3:
Hourly time.Ticker each hour





1. 
Add GetAllPoolMembersWithTimezones() to SQLiteStores.
Join with friend_pool_members, profiles, users, friend_pools to get:
pool_id, pool_mode, total_limit , profile_id & timezone
2. 
Add SecondsUntil4am(tz string, now time.Time) in pkg/store/ttl.go / valkey.go
returns next4am always
3.
Change initValkey to:
- use SET k,v NX EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds (a positive integer).
Shared pool:
- Use timezone of first user to join the pool. ( Using created_at)
Borrow pool:
- Each user gets their own timezone done
4.
Add KEEPTTL to set calls in initValkey + INCR/DECR
5. Add: ( too store/store.go) - pool methods
ResetShared(ctx context.Context, poolID string, limit int64, ttlSeconds int64) error
ResetBorrow(ctx context.Context, poolID, profileID string, limit int64, ttlSeconds int64) error

These are SET k,v, EX ttlSeconds, no NX

6.
pkg/store/resetter.go:
Resetter has valkey.Client / Pool store
Reference to DB to call GetAllPoolMembersWithTimezones
exports StartREsetJob(ctx context.Context, pool Pool, db *db.SqliteStores)
Spawns gorouting with time.NewTicker(time.Hour)

On tick:
1. GetAllPoolMembersWithTimezones() to get current pool state
2. for each borrow member: compute SecondsUntil4am for their timezone. If the key is missing from Valkey (i.e., TTL expired, key gone), call ResetBorrow(ctx, poolID, profileID, totalLimit, ttlSeconds).
3. For each shared pool: using the first person timezone, if the key is missing, call ResetShared.
To check if a key is missing: use EXISTS or GET — if it returns nil/0, re-seed it.

Use existance check for the ttl keys, as valkey handles expiry


7.
Add uh go store.StartResetJob(ctx, poolCacheStore, stores) to pkg/server/server.go


8.
Add memory store fallback for ResetShared & ResetBorrow 
