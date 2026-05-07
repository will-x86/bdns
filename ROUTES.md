This will contain routes and where they lie.


It'll be:


./api/ - web api + gRPC communication to DNS in ./dns/


API Routes:

# Authentication
Method	Path	Description
POST	/api/auth/signup	Create user with timezone, returns user ID
POST	/api/auth/login/{userID}	Login with user ID (returns user ID as confirmation)

# User
Method	Path	Description
GET	/api/users/{userID}	Get user info (timezone, created_at)
PUT	/api/users/{userID}	Update user (e.g., timezone)

# Profiles (child of User)
Method	Path	Description
GET	/api/users/{userID}/profiles	List all profiles for user
POST	/api/users/{userID}/profiles	Create profile (name)
GET	/api/profiles/{profileID}	Get profile details
PUT	/api/profiles/{profileID}	Update profile (name)
DELETE	/api/profiles/{profileID}	Delete profile

# Whitelists (per Profile)
Method	Path	Description
GET	/api/profiles/{profileID}/whitelist	List permanent whitelists
POST	/api/profiles/{profileID}/whitelist	Add permanent whitelist (domain)
DELETE	/api/profiles/{profileID}/whitelist/{domain}	Remove permanent whitelist
GET	/api/profiles/{profileID}/whitelist/temp	List temporary whitelists
POST	/api/profiles/{profileID}/whitelist/temp	Add temporary whitelist (domain, expires_at)
DELETE	/api/profiles/{profileID}/whitelist/temp/{domain}	Remove temporary whitelist

# Category Blocks (per Profile)
Method	Path	Description
GET	/api/profiles/{profileID}/categories	List blocked categories
POST	/api/profiles/{profileID}/categories	Block a category
DELETE	/api/profiles/{profileID}/categories/{category}	Unblock a category

# Time Blocks (per Profile)
Method	Path	Description
GET	/api/profiles/{profileID}/timeblocks	List time blocks
POST	/api/profiles/{profileID}/timeblocks	Create time block (category, start/end time, day)
DELETE	/api/profiles/{profileID}/timeblocks/{blockID}	Delete time block

# Friend Pools
Method	Path	Description
GET	/api/pools	List pools for user
POST	/api/pools	Create pool (name, mode: shared/borrow, limit)
GET	/api/pools/{poolID}	Get pool details
DELETE	/api/pools/{poolID}	Delete pool (owner only)
POST	/api/pools/{poolID}/join	Join pool (profileID)
POST	/api/pools/{poolID}/leave	Leave pool (profileID)
GET	/api/pools/{poolID}/members	List pool members
GET	/api/pools/{poolID}/blocks	List pool category blocks
POST	/api/pools/{poolID}/blocks	Block category for pool
DELETE	/api/pools/{poolID}/blocks/{category}	Unblock category for pool
GET	/api/pools/{poolID}/credits	Get remaining credits

# Blocklist Categories (read-only reference)
Method	Path	Description
GET	/api/categories	List available categories (fakenews, gambling, porn, social, unified)