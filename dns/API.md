# Management API

JSON API (Fiber v3) for driving config from a UI. Runs in the same binary on its
own port. Enable with `API_PORT` (or `API_ADDR=127.0.0.1:8080`).

Auth: `Authorization: Bearer <api_token>` on everything except `GET /health` and
`POST /users` (bootstrap — returns the token once). Requests are scoped to the
token's user.

```
POST   /api/v1/users                 {timezone} -> {id, api_token, ...}
GET    /api/v1/me   PATCH /me  DELETE /me   POST /me/token
GET    /api/v1/friends   POST /friends {friend_id}   DELETE /friends/:id
GET    /api/v1/categories
CRUD   /api/v1/profiles[/:pid]                        {name}
       /profiles/:pid/whitelist/permanent[/:domain]  {domain}
       /profiles/:pid/whitelist/temporary[/:domain]  {domain, ttl_seconds|expires_at}
       /profiles/:pid/category-blocks[/:category]     {category}
       /profiles/:pid/time-blocks                     {category, start_time, end_time, day}   # slots 0..95, day 0..7
CRUD   /api/v1/pools[/:poolID]                        {name, pool_mode, total_limit}
       /pools/:poolID/members[/:profileID]            {profile_id}   # creator only; must be own/friend profile
       /pools/:poolID/category-blocks[/:category]     {category}
GET    /api/v1/pools/:poolID/limits                   # live remaining; null until next daily reset
```

Pool reads need creator-or-member; pool mutations need creator. A profile `id` is
the DoT SNI / DoH subdomain a client uses when querying the proxy.
