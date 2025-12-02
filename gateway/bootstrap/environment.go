package bootstrap

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Env struct {
	Server       Server
	PrimaryRedis Redis
	OTP          OTP
	SMSGateway   SMSGateway
	Minio        Minio
	Postgres     Postgres
	Logger       Logger
	JWT          JWT
	Admin        Admin
}

type Admin struct {
	PhoneNumber string
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
	BackdoorCode string
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

type Logger struct {
	ConsoleOutput string
	LogFile       string
}

type JWT struct {
	AccessExpTime  time.Duration
	RefreshExpTime time.Duration
}

func NewEnvironment() *Env {
	godotenv.Load(".env")
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
			ExpiryMinute: getEnvInt("OTP_EXPIRATION_MINUTES", 5),
			MaxAttempts:  getEnvInt("OTP_MAX_ATTEMPTS", 5),
			BackdoorCode: os.Getenv("OTP_BACKDOOR_CODE"),
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
		Logger: Logger{
			ConsoleOutput: getEnvString("CONSOLE_OUTPUT", "true"),
			LogFile:       os.Getenv("LOG_FILE"),
		},
		JWT: JWT{
			AccessExpTime:  time.Duration(getEnvInt("JWT_ACCESS_EXP_MIN", 15)) * time.Minute,
			RefreshExpTime: time.Duration(getEnvInt("JWT_REFRESH_EXP_HOURS", 24*7)) * time.Hour,
		},
		Admin: Admin{
			PhoneNumber: os.Getenv("ADMIN_PHONE_NUMBER"),
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
