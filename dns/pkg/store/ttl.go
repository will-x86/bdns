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
	local := now.In(location)
	var target time.Time
	if local.Hour() >= 4 {
		target = time.Date(local.Year(), local.Month(), local.Day()+1, 4, 0, 0, 0, location)
	} else {
		target = time.Date(local.Year(), local.Month(), local.Day(), 4, 0, 0, 0, location)
	}
	return int(target.Sub(now).Seconds())
}
