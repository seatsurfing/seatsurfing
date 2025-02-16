package util

import "time"

const JsDateTimeFormat string = "2006-01-02T15:04:05"
const JsDateTimeFormatWithTimezone string = "2006-01-02T15:04:05-07:00"

func ParseJSDate(s string) (time.Time, error) {
	return time.Parse(JsDateTimeFormat, s)
}

func ToJSDate(date time.Time) string {
	return date.Format(JsDateTimeFormat)
}
