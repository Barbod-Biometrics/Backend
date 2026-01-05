package types

import (
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/date"
)

type JalaliDate time.Time

func (j *JalaliDate) UnmarshalParam(param string) error {
	if param == "" {
		return nil
	}

	gregorianTime, err := date.ParseJalaliToGregorian(param)
	if err != nil {
		return err
	}

	*j = JalaliDate(gregorianTime)
	return nil
}

func (j JalaliDate) ToTime() time.Time {
	return time.Time(j)
}

func (j JalaliDate) IsZero() bool {
	return time.Time(j).IsZero()
}
