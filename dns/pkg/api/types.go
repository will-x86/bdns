package api

import "github.com/will-x86/bdns/dns/pkg/db/models"

// ---- request DTOs ----

type createUserReq struct {
	Timezone string `json:"timezone"`
}

type updateMeReq struct {
	Timezone string `json:"timezone"`
}

type createProfileReq struct {
	Name string `json:"name"`
}

type updateProfileReq struct {
	Name string `json:"name"`
}

type addFriendReq struct {
	FriendID string `json:"friend_id"`
}

type domainReq struct {
	Domain string `json:"domain"`
}

// tempWhitelistReq accepts either an absolute expiry (unix seconds) or a TTL.
type tempWhitelistReq struct {
	Domain     string `json:"domain"`
	ExpiresAt  int64  `json:"expires_at"`
	TTLSeconds int64  `json:"ttl_seconds"`
}

type categoryReq struct {
	Category string `json:"category"`
}

type timeBlockReq struct {
	Category  string `json:"category"`
	StartTime int    `json:"start_time"`
	EndTime   int    `json:"end_time"`
	Day       int    `json:"day"`
}

type createPoolReq struct {
	Name       string `json:"name"`
	PoolMode   string `json:"pool_mode"`
	TotalLimit int64  `json:"total_limit"`
}

type updatePoolReq struct {
	Name       string `json:"name"`
	TotalLimit int64  `json:"total_limit"`
}

type addMemberReq struct {
	ProfileID string `json:"profile_id"`
}

// ---- response DTOs ----

// userWithToken exposes the api_token, only on create and rotate.
type userWithToken struct {
	models.User
	APIToken string `json:"api_token"`
}

// memberLimit is a borrow-pool member's remaining allowance; nil if unknown.
type memberLimit struct {
	ProfileID string `json:"profile_id"`
	Remaining *int64 `json:"remaining"`
}

// poolLimitsResp is Remaining for shared pools, Members for borrow pools.
type poolLimitsResp struct {
	PoolID     string        `json:"pool_id"`
	Mode       string        `json:"mode"`
	TotalLimit int64         `json:"total_limit"`
	Remaining  *int64        `json:"remaining,omitempty"`
	Members    []memberLimit `json:"members,omitempty"`
}
