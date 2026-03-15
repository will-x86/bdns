### AI Disclosure

Tests will be written by AI/me, anything else will be written by me.

I'm not very familiar with "proper" tests in Go

# Bad DNS 


The idea here is a social based DNS, where friends share credits / queries.

The goal is not to be a full blown DNS, but rather a proxy that adds some features


Components: 
- dns - Core DNS using go, goal is to do the following:
    - Parse query ( extract user_id via via sni in DoT/ or subdomain in DoH )
    - Do a simple check if user has the ability to check said side ( credits / full block / whatever)
    - Some user interface..?





### Friend based limits
The idea is, you invite a friend, and you both have say ~6k queries for X category per day. Primarily social media based focus.


### Rules engine flow: ( Some not implemented)
- All rules are per profile unless specified otherwise
- Is this domain permanently whitelisted -> per profile # Primarily for "I *need* this domain"
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
### Deploying 

// see [dns/DEPLOY.md](/dns/DEPLOY.md)



### Test locally 
```bash
openssl genrsa -out server.key 2048\
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650\
```
add to .env:
```
PORT=8533
KEY_PATH=./server.key
CRT_PATH=./server.crt
GOOSE_MIGRATION_DIR=./migrations/
```

```bash
go run ./... -ingest
go run ./... -seed
kdig @127.0.0.1 -p 8533 +tls-sni=bobby google.com # refused
kdig @127.0.0.1 -p 8533 +tls-sni=aabbccdd.dns.example.com google.com # accepted
```
-ingest downloads all blocklists then exists
-seed runs after executing the content in seed.sql ( create init user)


















