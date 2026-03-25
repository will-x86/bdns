package store

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

func SecondsUntil4AM(ctx context.Context, tz string, now time.Time) int {
	log := zerolog.Ctx(ctx).With().Str("component", "secondsuntil4am").Logger()
	location, err := time.LoadLocation(tz)
	if err != nil {
		// Should *never* fail
		log.Fatal().Err(err).Str("timezone", tz).Msg("users timezone is wrong")
	}
	now = now.In(location)
	if now.Hour() >= 4 { // after or in 4am
		return int(time.Until(time.Date(now.Year(), now.Month(), now.Day()+1, 4, 0, 0, 0, location)).Seconds())
	} else {
		return int(time.Until(time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, location)).Seconds())
	}
}
