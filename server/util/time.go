package util

import "time"

// SameWallClock compares a and b ignoring their time zones, matching how
// TIMESTAMP without time zone columns store them.
func SameWallClock(a, b time.Time) bool {
	const layout = "2006-01-02 15:04:05.999999"
	return a.Format(layout) == b.Format(layout)
}
