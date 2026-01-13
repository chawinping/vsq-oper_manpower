package timezone

import (
	"time"
)

const ThailandTimeZone = "Asia/Bangkok"

// GetThailandTime returns current time in Thailand timezone
func GetThailandTime() time.Time {
	loc, _ := time.LoadLocation(ThailandTimeZone)
	return time.Now().In(loc)
}

// ParseThailandTime parses a time string and returns time in Thailand timezone
func ParseThailandTime(layout, value string) (time.Time, error) {
	loc, err := time.LoadLocation(ThailandTimeZone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, loc)
}

// FormatThailandTime formats time in Thailand timezone
func FormatThailandTime(t time.Time, layout string) string {
	loc, _ := time.LoadLocation(ThailandTimeZone)
	return t.In(loc).Format(layout)
}






