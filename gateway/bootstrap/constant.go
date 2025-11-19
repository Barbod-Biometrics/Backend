package bootstrap

import (
	"fmt"
)

type Constants struct {
	RedisKey     RedisKey
	SMSTemplates SMSTemplates
}

type SMSTemplates struct {
	OTP string
}

type RedisKey struct {
}

func NewConstants() *Constants {
	return &Constants{
		SMSTemplates: SMSTemplates{
			OTP: "sendOTPTemplate",
		},
	}
}

func (r *RedisKey) GenerateOTPKey(value string) string {
	return fmt.Sprintf("otp:%s", value)
}
