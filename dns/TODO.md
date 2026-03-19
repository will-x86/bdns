- cache/ limits must go in uh FriendshipStore / limit store
- Rcache DoaminDNSCacheKey 
    - Swap to memory if no valkey 
    - Shared mode : pool:{pool_id}:credits = total_limit
    - Borrow mode : pool:{pool_id}:{profile_id}:credits = total_limit
    - Add DOmainDNSCachekey for shared + borrow
    - Decrement + set ( ttl .. ? ) 


- Add new rule for it + Stores , for each mode
- Add types
- Get stuff in DB 
- Pass down info in Stores 
- Limit popol



- in handler.go, add uh :
if profile is nil, check if it's a pool_id via ~~friend_pool_members~~:
    via h.stores.Pool.PoolExists(ctx,id string) // check in mem / redis


// two stores, PoolCacheStore -> redis / memory 
// PoolDBStore -> Sqlite database

