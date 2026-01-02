package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
)

func ParseJalaliToGregorian(jalaliDate string) (time.Time, error) {
	// cleaning
	jalaliDate = strings.TrimSpace(jalaliDate)
	if jalaliDate == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// Manual Parsing
	parts := strings.Split(jalaliDate, "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid format, expected YYYY-MM-DD")
	}

	y, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year: %v", err)
	}

	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month: %v", err)
	}

	d, err := strconv.Atoi(parts[2])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day: %v", err)
	}

	// Construct the Persian Date object
	pt := ptime.Date(y, ptime.Month(m), d, 0, 0, 0, 0, ptime.Iran())

	// Return standard Go Time (Gregorian)
	return pt.Time(), nil
}

func ToJalaliString(t time.Time) string {
	// Convert standard time to Persian time wrapper
	pt := ptime.New(t)
	// Format it
	return pt.Format("yyyy-MM-dd HH:mm:ss")
}
