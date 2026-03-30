package test

import (
	"testing"
	"time"

	. "github.com/seatsurfing/seatsurfing/server/testutil"
	. "github.com/seatsurfing/seatsurfing/server/util"
)

func TestAttachSomeTimezoneMultipleTimes(t *testing.T) {
	time := time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC)
	timeString1 := time.String()

	time, _ = AttachTimezoneInformationTz(time, "Europe/Berlin")
	timeString2 := time.String()
	CheckTestBool(t, true, timeString1 != timeString2)

	time, _ = AttachTimezoneInformationTz(time, "Europe/Berlin")
	timeString3 := time.String()
	CheckTestBool(t, true, timeString2 == timeString3)
}
