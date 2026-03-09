# Bad DNS 


The idea here is a social based DNS, where friends share credits / queries.

The goal is not to be a full blown DNS, but rather a beautiful proxy to learn some technologies.


Components: ( tood most LMFAO ) 

- dns - Core DNS using go, goal is to do the following:
    - Parse query ( extract user_id via via sni in DoT/ or subdomain in DoH )
    - Do a simple check if user has the ability to check said side ( credits / full block / whatever)
    - Provide smoochy api for the app





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
- [ ] Redis TTL cache for resolved domains
- [ ] SQLite schema: domains, categories, block_rules
- [ ] Blocklist ingestion (import from community lists e.g. Steven Black, OISD)
- [ ] SNI -> user_id parsing for DoT 
- [ ] Basic allow/block rule engine


## Blocking rules 
- [ ] User profiles - each user has profiles, e.g. "laptop" / "phone"
- [ ] Category-based blocking: Social Media, Gaming, Adult, News, etc.
- [ ] Time-of-day rules ("no TikTok after 9pm")
- [ ] Schedule blocks ( "Work day,9am-5pm block Facebook" )
- [ ] Whitelist overrides
- [ ] Per-profile DNS query logging

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
