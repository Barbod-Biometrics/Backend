package bootstrap

import (
	"fmt"
)

type Constants struct {
	RedisKey     RedisKey
	SMSTemplates SMSTemplates
	JWTKeysPath  JWTKeysPath
	Field        ErrorField
	Tag          ErrorTag
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

type ErrorField struct {
	User                string
	Phone               string
	Email               string
	Password            string
	OTP                 string
	RegistrationNumber  string
}

type ErrorTag struct {
	AlreadyRegistered      string
	MinimumLength          string
	ContainsLowercase      string
	ContainsUppercase      string
	ContainsNumber         string
	ContainsSpecialChar    string
	Expired                string
	Invalid                string
	NotRegistered          string
	NotVerified            string
	NotActive              string
	InvalidAuthCredentials string
	ExpiredAuthToken       string
	InvalidAuthToken       string
	Unauthorized           string
	NotExist               string
	AlreadyExist           string
	ForbiddenStatus        string
	InvalidNumber          string
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
		Field: ErrorField{
			User:                "user",
			Phone:               "phone",
			Email:               "email",
			Password:            "password",
			OTP:                 "otp",
		},
		Tag: ErrorTag{
			AlreadyRegistered:      "alreadyRegistered",
			MinimumLength:          "minimumLength",
			ContainsLowercase:      "containsLowercase",
			ContainsUppercase:      "containsUppercase",
			ContainsNumber:         "containsNumber",
			ContainsSpecialChar:    "containsSpecialChar",
			Expired:                "Expired",
			Invalid:                "invalid",
			NotRegistered:          "notRegistered",
			NotVerified:            "notVerified",
			InvalidAuthCredentials: "invalidAuthCredentials",
			ExpiredAuthToken:       "expiredAuthToken",
			InvalidAuthToken:       "invalidAuthToken",
			Unauthorized:           "unauthorized",
			NotExist:               "notExist",
			AlreadyExist:           "alreadyExist",
			ForbiddenStatus:        "forbiddenStatus",
			InvalidNumber:          "invalidNumber",
		},
	}
}

func (r *RedisKey) GenerateOTPKey(value string) string {
	return fmt.Sprintf("otp:%s", value)
}
