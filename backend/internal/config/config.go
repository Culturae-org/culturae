// backend/internal/config/config.go

package config

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	envTrue       = "true"
	envProduction = "production"
	envProd       = "prod"
	envDev        = "development"
)

type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error [%s]: %s", e.Field, e.Message)
}

type ValidationErrors []error

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return fmt.Sprintf("configuration validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

type Config struct {
	JWTSecret      string
	ServerPort     string
	Env            string
	AppMode        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	AllowedOrigins []string
	TrustedProxies []string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string

	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOBucketName string
	MinIOUseSSL     bool

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	CookieSecure bool

	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   time.Duration

	DatasetCheckInterval time.Duration

	SeedGameTemplates bool
}

const (
	AppModeFull    = "full"
	AppModeAdmin   = "admin"
	AppModeHeadless = "headless"
)

func Load() (*Config, error) {
	if os.Getenv("ENV") == "" || os.Getenv("ENV") == envDev || os.Getenv("ENV") == "dev" {
		_ = godotenv.Load()
	}

	useSSL := os.Getenv("MINIO_USE_SSL") == envTrue

	redisDB := 0
	if redisDBStr := os.Getenv("REDIS_DB"); redisDBStr != "" {
		if db, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = db
		}
	}

	readTimeout := parseDurationOrDefault("READ_TIMEOUT", 15*time.Second)
	writeTimeout := parseDurationOrDefault("WRITE_TIMEOUT", 15*time.Second)
	idleTimeout := parseDurationOrDefault("IDLE_TIMEOUT", 60*time.Second)

	allowedOrigins := parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS"))

	rateLimitEnabled := getEnvOrDefault("RATE_LIMIT_ENABLED", "false") == envTrue
	rateLimitRequests := parseIntOrDefault("RATE_LIMIT_REQUESTS", 200)
	rateLimitWindow := parseDurationOrDefault("RATE_LIMIT_WINDOW", time.Minute)

	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))

	appMode := getEnvOrDefault("APP_MODE", AppModeFull)
	if appMode != AppModeFull && appMode != AppModeAdmin && appMode != AppModeHeadless {
		return nil, &ConfigError{Field: "APP_MODE", Message: fmt.Sprintf("must be one of: %s, %s, %s", AppModeFull, AppModeAdmin, AppModeHeadless)}
	}

	cfg := &Config{
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnvOrDefault("POSTGRES_PORT", "5432"),
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "culturae"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", "culturae"),
		PostgresDBName:   getEnvOrDefault("POSTGRES_DB_NAME", "culturae"),

		JWTSecret:      getEnvOrDefault("JWT_SECRET", "dev_jwt_secret_change_in_production"),
		ServerPort:     getEnvOrDefault("PORT", "8080"),
		Env:            getEnvOrDefault("ENV", envDev),
		AppMode:        appMode,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		AllowedOrigins: allowedOrigins,

		MinIOEndpoint:   getEnvOrDefault("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:  getEnvOrDefault("MINIO_ROOT_USER", "minioadmin"),
		MinIOSecretKey:  getEnvOrDefault("MINIO_ROOT_PASSWORD", "minioadmin"),
		MinIOBucketName: getEnvOrDefault("MINIO_BUCKET_NAME", "culturae"),
		MinIOUseSSL:     useSSL,

		RedisHost:     getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort:     getEnvOrDefault("REDIS_PORT", "6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       redisDB,

		CookieSecure:      os.Getenv("COOKIE_SECURE") == envTrue,
		RateLimitEnabled:  rateLimitEnabled,
		RateLimitRequests: rateLimitRequests,
		RateLimitWindow:   rateLimitWindow,

		DatasetCheckInterval: parseDurationOrDefault("DATASET_CHECK_INTERVAL", 12*time.Hour),
		SeedGameTemplates:    getEnvOrDefault("SEED_GAME_TEMPLATES", envTrue) == envTrue,
		TrustedProxies:       trustedProxies,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) GetPostgresDSN() string {
	sslMode := "disable"
	if c.IsProduction() {
		sslMode = getEnvOrDefault("POSTGRES_SSLMODE", "require")
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDBName, sslMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func parseDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func parseAllowedOrigins(originsStr string) []string {
	if originsStr == "" {
		return []string{"http://localhost:3000"}
	}

	origins := strings.Split(originsStr, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

func parseTrustedProxies(proxiesStr string) []string {
	if proxiesStr == "" {
		return nil
	}

	proxies := strings.Split(proxiesStr, ",")
	var result []string
	for _, proxy := range proxies {
		if trimmed := strings.TrimSpace(proxy); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

func (c *Config) IsOriginAllowed(origin string) bool {
	for _, o := range c.AllowedOrigins {
		if o == "*" {
			return true
		}
	}
	return slices.Contains(c.AllowedOrigins, origin)
}

func (c *Config) IsProduction() bool {
	return c.Env == envProduction || c.Env == envProd
}

func (c *Config) IsDevelopment() bool {
	return c.Env == envDev || c.Env == "dev"
}

func (c *Config) IsAdmin() bool {
	return c.AppMode == AppModeAdmin
}

func (c *Config) IsHeadless() bool {
	return c.AppMode == AppModeHeadless
}

func (c *Config) IsFullMode() bool {
	return c.AppMode == AppModeFull
}

func (c *Config) Validate() error {
	var errs ValidationErrors

	if c.ServerPort == "" {
		errs = append(errs, &ConfigError{Field: "PORT", Message: "server port is required"})
	}

	if c.PostgresHost == "" {
		errs = append(errs, &ConfigError{Field: "POSTGRES_HOST", Message: "database host is required"})
	}
	if c.PostgresDBName == "" {
		errs = append(errs, &ConfigError{Field: "POSTGRES_DB_NAME", Message: "database name is required"})
	}
	if c.PostgresUser == "" {
		errs = append(errs, &ConfigError{Field: "POSTGRES_USER", Message: "database user is required"})
	}

	if c.RedisHost == "" {
		errs = append(errs, &ConfigError{Field: "REDIS_HOST", Message: "redis host is required"})
	}

	if c.JWTSecret == "" {
		errs = append(errs, &ConfigError{Field: "JWT_SECRET", Message: "JWT secret is required"})
	}

	if c.IsProduction() {
		if c.PostgresPassword == "" {
			errs = append(errs, &ConfigError{
				Field:   "POSTGRES_PASSWORD",
				Message: "database password is required in production",
			})
		}

		if len(c.JWTSecret) < 32 {
			errs = append(errs, &ConfigError{
				Field:   "JWT_SECRET",
				Message: "JWT secret must be at least 32 characters in production",
			})
		}

		if c.RedisPassword == "" {
			errs = append(errs, &ConfigError{
				Field:   "REDIS_PASSWORD",
				Message: "redis password is required in production",
			})
		}

		if !c.CookieSecure {
			errs = append(errs, &ConfigError{
				Field:   "COOKIE_SECURE",
				Message: "cookies must be secure in production",
			})
		}

		if c.MinIOEndpoint != "" {
			if c.MinIOAccessKey == "" || c.MinIOSecretKey == "" {
				errs = append(errs, &ConfigError{
					Field:   "MINIO",
					Message: "MinIO credentials are required when endpoint is configured",
				})
			}
		}

		for _, origin := range c.AllowedOrigins {
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				errs = append(errs, &ConfigError{
					Field:   "ALLOWED_ORIGINS",
					Message: fmt.Sprintf("localhost origins not allowed in production: %s", origin),
				})
				break
			}
		}
	}

	if c.RateLimitEnabled {
		if c.RateLimitRequests <= 0 {
			errs = append(errs, &ConfigError{
				Field:   "RATE_LIMIT_REQUESTS",
				Message: "rate limit requests must be positive",
			})
		}
		if c.RateLimitWindow <= 0 {
			errs = append(errs, &ConfigError{
				Field:   "RATE_LIMIT_WINDOW",
				Message: "rate limit window must be positive",
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
