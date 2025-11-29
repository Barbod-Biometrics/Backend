package bootstrap

import (
	"fmt"
)

type Constants struct {
	RedisKey     RedisKey
	SMSTemplates SMSTemplates
	JWTKeysPath  JWTKeysPath
}

type SMSTemplates struct {
	OTP string
}

type JWTKeysPath struct {
	PublicKey  string
	PrivateKey string
}

type RedisKey struct {
}

func NewConstants() *Constants {
	return &Constants{
		SMSTemplates: SMSTemplates{
			OTP: "sendOTPTemplate",
		},
		JWTKeysPath: JWTKeysPath{
			PublicKey:  "./internal/infrastructure/jwt/publicKey.pem",
			PrivateKey: "./internal/infrastructure/jwt/privateKey.pem",
		},
	}
}

func (r *RedisKey) GenerateOTPKey(value string) string {
	return fmt.Sprintf("otp:%s", value)
}
