# Bad DNS 


The idea here is a social based DNS, where friends share credits / queries.

The goal is not to be a full blown DNS, but rather a beautiful proxy to learn some technologies.


Components: ( tood most LMFAO ) 

- dns - Core DNS using go, goal is to do the following:
    - Parse query ( extract user_id via via sni in DoT/ or subdomain in DoH )
    - Do a simple check if user has the ability to check said side ( credits / full block / whatever)
    - Provide smoochy api for the app





### Friend based limits
The idea is, you invite a friend, and you both have say ~6k queries for X category per day. Primarily social media based focus.


### Rules engine flow:
- Is this domain permanently whitelisted -> per profile # Primarily for "oops need this domain"
- Is this domain temporarily whitelisted -> end of day whitelist 
- Is this category fully blocked for this profile ? # Primarily for ADS/Porn/Gambling
- Does this user have hard time blocks for this category ( No social media after 10 regardless ) 
- Does the user have a friend ? 
    - Is the site in their blocked category ? ( *shared* list of blocked things ) # Primarily social media
    - What kinda of limit are they using ?Borrow/Shared 
        - Borrow:
            - Borrowing is User A&B get 3k requests each, if user A runs out they can request some limit temporarily off of user B ( resets EOD)
            - Has this user got any limit left ? 
            - If yes, limit--, otherwise block
        - Shared:
            - Shared is where user A&B have a pool of ~6k requests, if user A uses 6k, neither user can use limits.
            - Has this user got any limit left in the pool? 
- Otherwise .. allow 
### Gen certs
```bash
openssl genrsa -out server.key 2048
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```



#### Testing SNI stuff

```bash
 kdig @127.0.0.1 -p 8533 +tls-sni=bobby google.com # Should print bobby
```


# Roadmap 


## DNS Server
- [x] Stand up a basic DNS resolver 
- [x] DoT Support
- [ ] DoH Support
- [x] Forward uncached queries upstream (1.1.1.1 / 8.8.8.8)
- [x] Redis TTL cache for resolved domains
- [x] SQLite schema: domains, categories, block_rules
- [-] Blocklist ingestion (import from community lists e.g. Steven Black, OISD)
- [ ] SNI -> user_id parsing for DoT 
- [ ] Basic allow/block rule engine


## Blocking rules 
- [ ] - Is this domain permanently whitelisted -> per profile # Primarily for "oops need this domain"
^^ SQLite 
- [ ] - Is this domain temporarily whitelisted -> end of day whitelist 
^^ SQLite
- [ ] - Is this category fully blocked for this profile ? # Primarily for ADS/Porn/Gambling
^^ SQLite
- [ ] - Does this user have hard time blocks for this category ( No social media after 10 regardless ) 
^^ SQLite
- [ ] - Does the user have a friend ? 
^^ SQLite 
- [ ]    - Is the site in their blocked category ? ( *shared* list of blocked things ) # Primarily social media
^^ SQLite
- [ ]    - What kinda of limit are they using ?Borrow/Shared 
- [ ]        - Borrow:
- [ ]            - Borrowing is User A&B get 3k requests each, if user A runs out they can request some limit temporarily off of user B ( resets EOD)
- [ ]            - Has this user got any limit left ? 
- [ ]            - If yes, limit--, otherwise block
^^Redis
- [ ]        - Shared:
- [ ]            - Shared is where user A&B have a pool of ~6k requests, if user A uses 6k, neither user can use limits.
- [ ]            - Has this user got any limit left in the pool? 
^^Redis
- [ ] - Otherwise .. allow 

## Credits 
- [ ] Redis-backed credit counters (reset at *4am* local time)
- [ ] Credit buckets per category per profile
- [ ] DNS resolution checks remaining credits before allowing
- [ ] Credits decrement on each request (or per-minute via heartbeat)
- [ ] block page (DNS redirect to block page)
- [ ] Warning page certificate

## Friends / Shared limits
- [ ]  Friend system: invite by code 
- [ ]  Friend system: invite by email.. 
- [ ]  Shared credit pools — friends share a daily budget together
- [ ]  "Accountability pairs" — if you overspend, friend gets notified
- [ ]  Friend can "lend" credits for 1 day, 3k each, lend 1k for a day to friend
- [ ]  Mutual block pacts: "we both agreed to block Reddit until 6pm"


## Web UI & dash
- [ ] Auth - sqlite
- [ ] Dashboard: today's usage by category, credits remaining
- [ ] Rule builder UI (drag-and-drop schedule blocks)
- [ ] Friends panel: shared pools, lending, pacts
- [ ] Query log explorer (filterable, searchable)
- [ ] Mobile-friendly 
- [ ] DNS setup wizard 



## Deployment idk 
- [ ] Single docker-compose.yml: DNS + API + UI + Redis + SQLite/Postgres
- [ ] .env-based config (upstream DNS, ports, secret key)
- [ ] Automatic blocklist updates (cron)
- [ ] Backup/restore for SQLite
- [ ] Developer docs


## Extra idk 
- [ ] DNSSEC Validation
- [ ] Mobile app
