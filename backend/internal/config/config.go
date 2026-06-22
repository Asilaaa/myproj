package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host string
	Port string

	FrontendOrigin string

	DatabaseURL string

	OpenAIAPIKey string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	OryURL string

	MaxUploadSizeBytes       int64
	AllowedUploadTypes       []string
	AIGuardrailsEnabled      bool
	AIMaxDescriptionChars    int
	UploadRateLimitPerMinute int
	ImportRateLimitPerMinute int
}

func Load() (*Config, error) {
	minIOUseSSL, err := getBoolEnv("MINIO_USE_SSL", false)
	if err != nil {
		return nil, err
	}

	aiGuardrailsEnabled, err := getBoolEnv("AI_GUARDRAILS_ENABLED", true)
	if err != nil {
		return nil, err
	}

	maxUploadSizeBytes, err := getInt64Env("MAX_UPLOAD_SIZE_BYTES", 10<<20)
	if err != nil {
		return nil, err
	}

	aiMaxDescriptionChars, err := getIntEnv("AI_MAX_DESCRIPTION_CHARS", 500)
	if err != nil {
		return nil, err
	}

	uploadRateLimitPerMinute, err := getIntEnv("UPLOAD_RATE_LIMIT_PER_MINUTE", 10)
	if err != nil {
		return nil, err
	}

	importRateLimitPerMinute, err := getIntEnv("IMPORT_RATE_LIMIT_PER_MINUTE", 20)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Host:           getEnv("HOST", "0.0.0.0"),
		Port:           getEnv("PORT", "8080"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "http://localhost:3000"),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),

		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucket:    getEnv("MINIO_BUCKET", ""),
		MinIOUseSSL:    minIOUseSSL,

		OryURL: getEnv("ORY_URL", ""),

		MaxUploadSizeBytes:       maxUploadSizeBytes,
		AllowedUploadTypes:       getListEnv("ALLOWED_UPLOAD_TYPES", []string{"image/jpeg", "image/png", "image/webp"}),
		AIGuardrailsEnabled:      aiGuardrailsEnabled,
		AIMaxDescriptionChars:    aiMaxDescriptionChars,
		UploadRateLimitPerMinute: uploadRateLimitPerMinute,
		ImportRateLimitPerMinute: importRateLimitPerMinute,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	required := map[string]string{
		"HOST":             c.Host,
		"PORT":             c.Port,
		"DATABASE_URL":     c.DatabaseURL,
		"OPENAI_API_KEY":   c.OpenAIAPIKey,
		"MINIO_ENDPOINT":   c.MinIOEndpoint,
		"MINIO_ACCESS_KEY": c.MinIOAccessKey,
		"MINIO_SECRET_KEY": c.MinIOSecretKey,
		"MINIO_BUCKET":     c.MinIOBucket,
		"ORY_URL":          c.OryURL,
	}

	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}

	if len(c.AllowedUploadTypes) == 0 {
		return fmt.Errorf("ALLOWED_UPLOAD_TYPES must contain at least one content type")
	}
	if c.AIMaxDescriptionChars <= 0 {
		return fmt.Errorf("AI_MAX_DESCRIPTION_CHARS must be greater than zero")
	}
	if c.UploadRateLimitPerMinute <= 0 {
		return fmt.Errorf("UPLOAD_RATE_LIMIT_PER_MINUTE must be greater than zero")
	}
	if c.ImportRateLimitPerMinute <= 0 {
		return fmt.Errorf("IMPORT_RATE_LIMIT_PER_MINUTE must be greater than zero")
	}

	return nil
}

func getEnv(key string, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func getBoolEnv(key string, defaultValue bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a valid boolean: %w", key, err)
	}

	return parsed, nil
}

func getIntEnv(key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func getInt64Env(key string, defaultValue int64) (int64, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func getListEnv(key string, defaultValue []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return append([]string(nil), defaultValue...)
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		normalized := strings.ToLower(strings.TrimSpace(part))
		if normalized == "" {
			continue
		}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return append([]string(nil), defaultValue...)
	}
	return result
}
