package util

import (
	"math/rand/v2"
	"strings"
	"time"
)

const JsDateTimeFormat string = "2006-01-02T15:04:05"
const JsDateTimeFormatWithTimezone string = "2006-01-02T15:04:05-07:00"

func ParseJSDate(s string) (time.Time, error) {
	return time.Parse(JsDateTimeFormat, s)
}

func ToJSDate(date time.Time) string {
	return date.Format(JsDateTimeFormat)
}

func MaxOf(vars ...int) int {
	max := vars[0]

	for _, i := range vars {
		if max < i {
			max = i
		}
	}

	return max
}

func GetDomainFromEmail(email string) string {
	mailParts := strings.Split(email, "@")
	if len(mailParts) != 2 {
		return ""
	}
	domain := strings.ToLower(mailParts[1])
	return domain
}

func GetRandomNumber(min, max int) int {
	if min >= max {
		return min
	}
	rnd := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	return min + rnd.IntN(max-min)
}
