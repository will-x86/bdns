package rules

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"codeberg.org/will-x86/bdns/dns/pkg/rule"
	"github.com/rs/zerolog"
)

/*
Users have timeblocks with start & end time with 0->95 values ( 15 min interval in 24hrs of a day)
*/
type TimeBlockRule struct{}

func (r *TimeBlockRule) Name() string { return "timeblock_rule" }

func (r *TimeBlockRule) Evaluate(ctx context.Context, rctx *rule.RuleContext) (rule.Decision, error) {
	log := zerolog.Ctx(ctx).With().Str("component", "timeblock-rule").Logger()
	ctx = log.WithContext(ctx)
	log.Trace().Msg("entering evaluate on timezone rule ")
	// is this category currently fully blocked for the user
	// e.g. "block social media 9-12am
	// factors into account users timezzone

	// Flow should be:
	// Grab category, blocks ( user should already be initializated from creation of rule-context)
	category, err := rctx.GetCategory(ctx)
	if err != nil {
		return rule.Decision{}, err

	}
	// No category, therefore cannot have timeblock block
	if category == "" {
		return rule.PassThrough(), nil
	}
	blocks, err := rctx.Stores.TimeBlock.GetTimeBlocks(ctx, rctx.ProfileID, category)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return rule.PassThrough(), nil
		}
		return rule.Decision{}, err
	}
	location, err := time.LoadLocation(rctx.User.Timezone)
	if err != nil {
		// Should *never* fail
		log.Fatal().Err(err).Str("timezone", rctx.User.Timezone).Msg("users timezone is wrong")
	}
	intervalToTime := func(interval int) string {
		hours := interval / 4
		minutes := (interval % 4) * 15
		return fmt.Sprintf("%02d:%02d", hours, minutes)
	}
	// We're taking the current day and blocking it into 96 values (15 min intervals 0-95)
	// We want the current interval the user is in
	// This is hour[0,23] times four, plus minutes[0,59]/15 ( 4 per hour)
	// 04:28 -> 16 +1
	now := rctx.Now.In(location)
	intervalIn := now.Hour()*4 + now.Minute()/15
	// 0 = 00:00-00:15
	// 20 = 05:00-05:15
	for k := range blocks {
		if blocks[k].StartTime <= intervalIn && blocks[k].EndTime >= intervalIn {
			// User is time blocked
			// EndTime is +1 as end time is inclusive
			return rule.Blocked(fmt.Sprintf("time block on %s category is active from %s to %s",
				category, intervalToTime(blocks[k].StartTime), intervalToTime(blocks[k].EndTime+1))), nil
		}
	}
	return rule.Allowed("permanent whitelist"), nil
}
