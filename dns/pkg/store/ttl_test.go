package store

import (
	"context"
	"testing"
	"time"
)

// All tests use a fixed arbitrary date so they never depend on wall clock.
// Any date works; we use 2030-06-15 throughout.
var testDate = time.Date(2030, 6, 15, 0, 0, 0, 0, time.UTC)

func at(hour, minute int) time.Time {
	return time.Date(2030, 6, 15, hour, minute, 0, 0, time.UTC)
}

func TestSecondsUntil4AM_AfterFourAM_NextDay(t *testing.T) {
	// 10:00 UTC → next 4 AM is tomorrow: 18h = 64800s
	got := SecondsUntil4AM(context.Background(), "UTC", at(10, 0))
	want := 18 * 60 * 60
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_BeforeFourAM_SameDay(t *testing.T) {
	// 01:00 UTC → 4 AM is later today: 3h = 10800s
	got := SecondsUntil4AM(context.Background(), "UTC", at(1, 0))
	want := 3 * 60 * 60
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_ExactlyFourAM_NextDay(t *testing.T) {
	// 04:00 exactly — hour >= 4, so rolls to next day: 24h = 86400s
	got := SecondsUntil4AM(context.Background(), "UTC", at(4, 0))
	want := 24 * 60 * 60
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_JustBeforeFourAM(t *testing.T) {
	// 03:59 UTC → 1 minute until 4 AM = 60s
	got := SecondsUntil4AM(context.Background(), "UTC", at(3, 59))
	want := 60
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_Midnight(t *testing.T) {
	// 00:00 UTC → 4 hours until 4 AM = 14400s
	got := SecondsUntil4AM(context.Background(), "UTC", at(0, 0))
	want := 4 * 60 * 60
	if got != want {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_Timezone_NewYork(t *testing.T) {
	// 08:00 UTC on 2030-06-15 = 04:00 EDT (UTC-4 in June)
	// hour >= 4 in NY, so next 4 AM is tomorrow: 24h = 86400s
	now := at(8, 0)
	got := SecondsUntil4AM(context.Background(), "America/New_York", now)
	want := 24 * 60 * 60
	if got != want {
		t.Fatalf("New York: want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_Timezone_NewYork_BeforeFourAM(t *testing.T) {
	// 06:00 UTC = 02:00 EDT (UTC-4 in June) — before 4 AM, same day
	// 4 AM EDT = 08:00 UTC, so 2h = 7200s from now
	now := at(6, 0)
	got := SecondsUntil4AM(context.Background(), "America/New_York", now)
	want := 2 * 60 * 60
	if got != want {
		t.Fatalf("New York before 4 AM: want %d, got %d", want, got)
	}
}

func TestSecondsUntil4AM_AlwaysPositive(t *testing.T) {
	cases := []struct {
		hour int
		tz   string
	}{
		{0, "UTC"}, {1, "UTC"}, {3, "UTC"}, {4, "UTC"},
		{12, "UTC"}, {23, "UTC"},
		{0, "America/New_York"}, {5, "Asia/Tokyo"},
	}
	for _, tc := range cases {
		now := time.Date(2030, 6, 15, tc.hour, 0, 0, 0, time.UTC)
		got := SecondsUntil4AM(context.Background(), tc.tz, now)
		if got <= 0 {
			t.Errorf("hour=%d tz=%s: want positive, got %d", tc.hour, tc.tz, got)
		}
	}
}

func TestSecondsUntil4AM_NeverMoreThan24Hours(t *testing.T) {
	for _, hour := range []int{0, 1, 2, 3, 4, 5, 12, 23} {
		now := time.Date(2030, 6, 15, hour, 0, 0, 0, time.UTC)
		got := SecondsUntil4AM(context.Background(), "UTC", now)
		if got > 24*60*60 {
			t.Errorf("hour=%d: %d exceeds 24h", hour, got)
		}
	}
}
