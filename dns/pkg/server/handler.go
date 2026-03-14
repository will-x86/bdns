package server

import (
	"context"
	"log"
	"time"

	"github.com/will-x86/bdns/dns/pkg/parser"
	"github.com/will-x86/bdns/dns/pkg/rcache"
	"github.com/will-x86/bdns/dns/pkg/rule"
)

type handler struct {
	upstream  DNSUpstream
	cache     *rcache.Cache
	write     func([]byte) error
	engine    *rule.Engine
	stores    rule.Stores
	profileID string
}

func (h *handler) handle(ctx context.Context, requestBytes []byte, remoteAddr string) {
	log.Printf("Received request from %s\n", remoteAddr)

	msg := parser.Message()
	if err := msg.Parse(requestBytes); err != nil {
		log.Printf("Error parsing DNS message from %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Parsed request: %s", msg.Header.String())

	if len(msg.Questions) == 0 {
		log.Printf("Request from %s has no questions — dropping\n", remoteAddr)
		return
	}
	if len(msg.Questions) > 1 {
		log.Printf("Request from %s has %d questions\n", remoteAddr, len(msg.Questions))
	}

	q := msg.Questions[0]
	qtypeStr, ok := parser.TypeToString[q.QType]
	if !ok {
		qtypeStr = "UNKNOWN"
	}

	// Refuse non-authed users
	if h.profileID == "" {
		log.Printf("No SNI/profileID from %s — refusing\n", remoteAddr)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}
	if h.stores.Profile == nil {
		log.Printf("No profile store configured, refusing %s\n", remoteAddr)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}
	// Pre-grab things every user has, other things such as category
	profile, user, profileErr := h.stores.Profile.GetProfileWithUser(ctx, h.profileID)
	if profileErr != nil {
		log.Printf("DB error fetching profile %s: %v\n", h.profileID, profileErr)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}
	if profile == nil {
		log.Printf("Profile %s not found in DB\n", h.profileID)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}

	decision, ruleErr := h.engine.Evaluate(ctx, &rule.RuleContext{
		Domain:    q.QName,
		ProfileID: h.profileID,
		Now:       time.Now(),
		Profile:   profile,
		User:      user,
		Stores:    h.stores,
	})
	if ruleErr != nil {
		log.Printf("Rule engine error for %s %s: %v\n", q.QName, qtypeStr, ruleErr)
		// Fail open
	} else if decision.Verdict == rule.VerdictBlock {
		log.Printf("Blocked %s %s for profile %s: %s\n", q.QName, qtypeStr, h.profileID, decision.Reason)
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Printf("Error sending REFUSED to %s: %v\n", remoteAddr, err)
		}
		return
	}

	var (
		responseBytes []byte
		err           error
	)

	if h.cache != nil {
		responseBytes, err = h.cache.QueryRace(
			ctx,
			requestBytes,
			q.QName,
			qtypeStr,
			func(ctx context.Context) ([]byte, error) {
				return h.upstream.SendQuery(requestBytes)
			},
		)
	} else {
		responseBytes, err = h.upstream.SendQuery(requestBytes)
	}

	if err != nil {
		log.Printf("Error resolving %s %s: %v\n", q.QName, qtypeStr, err)
		return
	}

	if err := h.write(responseBytes); err != nil {
		log.Printf("Error sending response to client %s: %v\n", remoteAddr, err)
		return
	}
	log.Printf("Sent response to client %s\n", remoteAddr)

	resMsg := parser.Message()
	if err := resMsg.Parse(responseBytes); err != nil {
		log.Printf("Error parsing DNS response: %v\n", err)
		return
	}
	log.Printf("Parsed response %s", resMsg.Header.String())
}
