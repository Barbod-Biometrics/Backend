package bootstrap

import (
	"os"
	"strconv"
)

type Env struct {
	Server       Server
	PrimaryRedis Redis
	OTP          OTP
	SMSGateway   SMSGateway
	Minio        Minio
}

type Server struct {
	Port string
	Mode string
}

type Redis struct {
	Port      string
	Address   string
	Password  string
	RDBNumber string
}

type OTP struct {
	Length       int
	ExpiryMinute int
	MaxAttempts  int
}

type SMSGateway struct {
	APIKey string
}

type Minio struct {
	Port         string
	PanelPort    string
	Host         string
	UserRoot     string
	PasswordRoot string
}

func NewEnvironment() *Env {
	return &Env{
		Server: Server{
			Port: os.Getenv("SERVER_PORT"),
			Mode: os.Getenv("SERVER_MODE"),
		},
		PrimaryRedis: Redis{
			Port:      os.Getenv("RDB_PORT"),
			Address:   os.Getenv("RDB_ADDRESS"),
			Password:  os.Getenv("RDB_PASSWORD"),
			RDBNumber: os.Getenv("RDB_NUMBER"),
		},
		OTP: OTP{
			Length:       getEnvInt("OTP_LENGTH", 6),
			ExpiryMinute: getEnvInt("OTP_EXPIRY_MINUTES", 2),
			MaxAttempts:  getEnvInt("OTP_MAX_ATTEMPTS", 3),
		},
		SMSGateway: SMSGateway{
			APIKey: os.Getenv("SMS_GATEWAY_API_KEY"),
		},
		Minio: Minio{},
	}
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return defaultVal
}
