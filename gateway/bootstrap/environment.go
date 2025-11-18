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
	Postgres     Postgres
}

type Postgres struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}
type Server struct {
	Port string
	Mode string
}

type Redis struct {
	Port     string
	Address  string
	Password string
	DB       int
	PoolSize int
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
			Mode: getEnvString("SERVER_MODE", "debug"),
		},
		PrimaryRedis: Redis{
			Port:     os.Getenv("REDIS_PORT"),
			Address:  os.Getenv("REDIS_HOST"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		OTP: OTP{
			Length:       getEnvInt("OTP_LENGTH", 6),
			ExpiryMinute: getEnvInt("OTP_EXPIRY_MINUTES", 2),
			MaxAttempts:  getEnvInt("OTP_MAX_ATTEMPTS", 3),
		},
		SMSGateway: SMSGateway{
			APIKey: os.Getenv("SMS_GATEWAY_API_KEY"),
		},
		Minio: Minio{
			Port:         os.Getenv("MINIO_PORT"),
			PanelPort:    os.Getenv("MINIO_PANEL_PORT"),
			Host:         os.Getenv("MINIO_HOST"),
			UserRoot:     os.Getenv("MINIO_ROOT_USER"),
			PasswordRoot: os.Getenv("MINIO_ROOT_PASSWORD"),
		},
		Postgres: Postgres{
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
			DBName:   os.Getenv("POSTGRES_DB"),
		},
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

func getEnvString(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
