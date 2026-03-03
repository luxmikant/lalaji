package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Log      LogConfig
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port        string
	GinMode     string
	CORSOrigins string // comma-separated list of allowed origins
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	URL                string // DATABASE_URL takes precedence if set (Neon / Render)
	Host               string
	Port               string
	User               string
	Password           string
	DBName             string
	SSLMode            string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetimeMin time.Duration
}

// DSN returns the PostgreSQL connection string.
// If a full DATABASE_URL is configured it is returned directly;
// otherwise the individual parameters are assembled.
func (d DatabaseConfig) DSN() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL      string // REDIS_URL takes precedence if set (Render Redis)
	Host     string
	Port     string
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// parseRedisURL extracts host, port, password from a Redis URL.
// Supported formats: redis://:password@host:port or rediss://:password@host:port
func parseRedisURL(rawURL string) (host, port, password string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "localhost", "6379", ""
	}
	if u.User != nil {
		password, _ = u.User.Password()
	}
	h := u.Hostname()
	p := u.Port()
	if h == "" {
		h = "localhost"
	}
	if p == "" {
		p = "6379"
	}
	return h, p, password
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level string
}

// Load reads configuration from .env file and environment variables.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	// Allow .env to be missing (rely on env vars in production)
	_ = viper.ReadInConfig()

	// Build Redis config — prefer REDIS_URL (Render) over individual vars
	redisURL := strings.TrimSpace(viper.GetString("REDIS_URL"))
	var redisCfg RedisConfig
	if redisURL != "" {
		h, p, pw := parseRedisURL(redisURL)
		redisCfg = RedisConfig{
			URL:      redisURL,
			Host:     h,
			Port:     p,
			Password: pw,
			DB:       viper.GetInt("REDIS_DB"),
		}
	} else {
		redisCfg = RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:        viper.GetString("SERVER_PORT"),
			GinMode:     viper.GetString("GIN_MODE"),
			CORSOrigins: viper.GetString("CORS_ALLOWED_ORIGINS"),
		},
		Database: DatabaseConfig{
			URL:                strings.TrimSpace(viper.GetString("DATABASE_URL")),
			Host:               viper.GetString("DB_HOST"),
			Port:               viper.GetString("DB_PORT"),
			User:               viper.GetString("DB_USER"),
			Password:           viper.GetString("DB_PASSWORD"),
			DBName:             viper.GetString("DB_NAME"),
			SSLMode:            viper.GetString("DB_SSLMODE"),
			MaxOpenConns:       viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:       viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLifetimeMin: time.Duration(viper.GetInt("DB_CONN_MAX_LIFETIME_MINUTES")) * time.Minute,
		},
		Redis: redisCfg,
		JWT: JWTConfig{
			Secret:      viper.GetString("JWT_SECRET"),
			ExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}

	// Defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	// Render sets the PORT env var; honour it if SERVER_PORT not explicit
	if renderPort := strings.TrimSpace(viper.GetString("PORT")); renderPort != "" && viper.GetString("SERVER_PORT") == "" {
		cfg.Server.Port = renderPort
	}
	if cfg.Server.GinMode == "" {
		cfg.Server.GinMode = "debug"
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.ConnMaxLifetimeMin == 0 {
		cfg.Database.ConnMaxLifetimeMin = 5 * time.Minute
	}
	if cfg.Redis.Host == "" && cfg.Redis.URL == "" {
		cfg.Redis.Host = "localhost"
		cfg.Redis.Port = "6379"
	}

	return cfg, nil
}
