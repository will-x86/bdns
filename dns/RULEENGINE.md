### Rule engine




Idea behind this is to pencil out the rule engine for helping me code lol

## Blocking rules 
- [x] - Is this domain permanently whitelisted -> per profile # Primarily for "oops need this domain"
^^ SQLite 
- [x] - Is this domain temporarily whitelisted -> end of day whitelist 
^^ SQLite
- [x] - Is this category fully blocked for this profile ? # Primarily for ADS/Porn/Gambling
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



### Flow 

Perprofile whitelist -> temporary whitelist -> 
category block -> time block -> 
friend block {is in category} -> 
{borrow limit hit / shared limit hit} -> allow

### data needed at each step

#### All below require domain+user_id
Perprofile whitelist -> domain, user_id -> grab is_whitelisted from *table to be made* (ttbm)
Temporary whitelist -> domain, user_id -> grab data from sqlite (ttbm)
#### All below require category
Category block -> domain, category, user_id -> grab data from sqlite (ttbm)
Time block -> domain, category, user_id, time + users_timezone -> grab data from sqlite (ttbm)
Friend block -> domain, category




