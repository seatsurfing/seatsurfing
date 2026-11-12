package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestSameWallClockIgnoresTimeZone(t *testing.T) {
	berlin, _ := time.LoadLocation("Europe/Berlin")
	a := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	b := time.Date(2026, 11, 12, 9, 0, 0, 0, berlin)
	CheckTestBool(t, true, SameWallClock(a, b))
}

func TestSameWallClockDifferentTimes(t *testing.T) {
	a := time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC)
	CheckTestBool(t, false, SameWallClock(a, a.Add(time.Minute)))
	CheckTestBool(t, false, SameWallClock(a, a.Add(time.Microsecond)))
	CheckTestBool(t, false, SameWallClock(a, a.AddDate(0, 0, 1)))
}
