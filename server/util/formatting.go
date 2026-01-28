package util

import (
	"errors"
	"html"
	"math/rand/v2"
	"reflect"
	"regexp"
	"strconv"
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

func CSSDimensionsToPixels(value string) (float64, error) {
	const dpi = 96.0
	re := regexp.MustCompile(`^\s*([0-9]*\.?[0-9]+)\s*(px|pt|pc|mm|cm|in)?\s*$`)
	m := re.FindStringSubmatch(value)
	if m == nil {
		return 0, errors.New("invalid unit or format")
	}

	v, _ := strconv.ParseFloat(m[1], 64)
	unit := m[2]

	switch unit {
	case "", "px":
		return v, nil
	case "in":
		return v * dpi, nil
	case "cm":
		return v * dpi / 2.54, nil
	case "mm":
		return v * dpi / 25.4, nil
	case "pt":
		return v * dpi / 72.0, nil
	case "pc":
		return v * 16.0, nil
	default:
		return 0, errors.New("unsupported unit")
	}
}

func EscapeStringsInStruct(m interface{}) error {
	msValuePtr := reflect.ValueOf(m)
	msValue := msValuePtr.Elem()
	for i := 0; i < msValue.NumField(); i++ {
		field := msValue.Field(i)
		// Ignore fields that don't have the same type as a string
		if field.Type() != reflect.TypeOf("") {
			continue
		}
		str := field.Interface().(string)
		str = strings.TrimSpace(str)
		str = html.EscapeString(str)
		field.SetString(str)
	}
	return nil
}
