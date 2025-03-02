package util

import (
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
