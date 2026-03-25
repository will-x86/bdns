package server

import (
	"context"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/parser"
	"codeberg.org/will-x86/bdns/dns/pkg/rcache"
	"codeberg.org/will-x86/bdns/dns/pkg/rule"
	"github.com/rs/zerolog"
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
	log := zerolog.Ctx(ctx).With().Str("component", "handle").Logger()
	log.Debug().Msg("recevied request")

	msg := parser.Message()
	if err := msg.Parse(requestBytes); err != nil {
		log.Error().Err(err).Msg("error parsing DNS message")
		return
	}
	log.Trace().Any("message", msg).Msg("parsed request")

	if len(msg.Questions) == 0 {
		log.Warn().Msg("request has no questions, dropping")
		return
	}
	if len(msg.Questions) > 1 {
		log.Warn().Int("num-questions", len(msg.Questions)).Msg("request has multiple questions")
	}

	q := msg.Questions[0]
	qtypeStr, ok := parser.TypeToString[q.QType]
	if !ok {
		qtypeStr = "UNKNOWN"
	}

	// Refuse non-authed users
	if h.profileID == "" {
		log.Warn().Msg("no sni/profileID from user, refusing")
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Error().Err(err).Msg("error sending REFUSED")
		}
		return
	}
	if h.stores.Profile == nil {
		log.Warn().Msg("no profile store configured, refusing")
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Error().Err(err).Msg("error sending REFUSED")
		}
		return
	}
	// Pre-grab things every user has, other things such as category
	profile, user, profileErr := h.stores.Profile.GetProfileWithUser(ctx, h.profileID)
	if profileErr != nil {
		log.Error().Str("profile-id", h.profileID).Err(profileErr).Msg("DB error fetching profile")
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Error().Err(err).Msg("error sending REFUSED")
		}
		return
	}
	if profile == nil {
		log.Warn().Str("profile-id", h.profileID).Msg("profile not found in DB")
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Error().Err(err).Msg("error sending REFUSED")
		}
		return
	}
	log = log.With().
		Str("profile-id", h.profileID).
		Str("remote", remoteAddr).
		Logger()
	ctx = log.WithContext(ctx)

	decision, ruleErr := h.engine.Evaluate(ctx, &rule.RuleContext{
		Domain:    q.QName,
		ProfileID: h.profileID,
		Now:       time.Now(),
		Profile:   profile,
		User:      user,
		Stores:    h.stores,
	})
	if ruleErr != nil {
		log.Error().Err(ruleErr).Str("q-name", q.QName).Str("q-type", qtypeStr).Msg("rule engine error")
		// Fail open
	} else if decision.Verdict == rule.VerdictBlock {
		log.Info().
			Str("q-name", q.QName).
			Str("q-type", qtypeStr).
			Str("reason", decision.Reason).
			Msg("blocked query")
		if err := h.write(buildRefusedResponse(requestBytes)); err != nil {
			log.Error().Err(err).Msg("error sending REFUSED")
		}
		return
	}

	var (
		responseBytes []byte
		err           error
	)

	if h.cache != nil {
		responseBytes, err = h.cache.QueryRace(
			log.WithContext(ctx),
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
		log.Error().Err(err).Str("q-name", q.QName).Str("q-type", qtypeStr).Msg("error resolving query")
		return
	}
	if err := h.write(responseBytes); err != nil {
		log.Error().Err(err).Msg("error sending response")
		return
	}
	log.Debug().Msg("sent response")

	resMsg := parser.Message()
	if err := resMsg.Parse(responseBytes); err != nil {
		log.Warn().Err(err).Msg("error parsing DNS response")
		return
	}
	log.Trace().Str("header", resMsg.Header.String()).Msg("parsed response")
}
