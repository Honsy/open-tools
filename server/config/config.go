package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PostgresHost     string `yaml:"postgresHost"`
	PostgresPort     string `yaml:"postgresPort"`
	PostgresUser     string `yaml:"postgresUser"`
	PostgresPassword string `yaml:"postgresPassword"`
	PostgresDB       string `yaml:"postgresDB"`
	PostgresSSLMode  string `yaml:"postgresSSLMode"`
	AdminUser        string `yaml:"adminUser"`
	AdminPassword    string `yaml:"adminPassword"`
	SessionSecret    string `yaml:"sessionSecret"`
	AdminSecret      string `yaml:"adminSecret"`

	R2AccountID       string `yaml:"r2AccountId"`
	R2AccessKeyID     string `yaml:"r2AccessKeyId"`
	R2SecretAccessKey string `yaml:"r2SecretAccessKey"`
	R2Bucket          string `yaml:"r2Bucket"`
	R2Endpoint        string `yaml:"r2Endpoint"`
	R2PublicBase      string `yaml:"r2PublicBase"`

	SMTPHost     string `yaml:"smtpHost"`
	SMTPPort     string `yaml:"smtpPort"`
	SMTPUser     string `yaml:"smtpUser"`
	SMTPPassword string `yaml:"smtpPassword"`
	SMTPFrom     string `yaml:"smtpFrom"`

	AIBailianAPIKey     string `yaml:"aiBailianApiKey"`
	AIBailianChatModels string `yaml:"aiBailianChatModels"`
	DailyArticles       bool   `yaml:"dailyArticles"`
}

func Load() (Config, error) {
	cfg := Config{
		PostgresHost:    "127.0.0.1",
		PostgresPort:    "5432",
		PostgresUser:    "kaiwu",
		PostgresDB:      "opentools",
		PostgresSSLMode: "disable",
		AdminUser:       "admin",
	}
	path := "config.yaml"
	if _, err := os.Stat(path); err != nil {
		alt := filepath.Join("server", "config.yaml")
		if _, err2 := os.Stat(alt); err2 == nil {
			path = alt
		}
	}
	if raw, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return cfg, err
		}
	}
	cfg.PostgresHost = first(os.Getenv("KAIWU_PGHOST"), cfg.PostgresHost)
	cfg.PostgresPort = first(os.Getenv("KAIWU_PGPORT"), cfg.PostgresPort)
	cfg.PostgresUser = first(os.Getenv("KAIWU_PGUSER"), cfg.PostgresUser)
	cfg.PostgresDB = first(os.Getenv("KAIWU_PGDATABASE"), cfg.PostgresDB)
	cfg.PostgresSSLMode = first(os.Getenv("KAIWU_PGSSLMODE"), cfg.PostgresSSLMode)
	if password, ok := os.LookupEnv("KAIWU_PGPASSWORD"); ok {
		cfg.PostgresPassword = password
	}
	cfg.AdminUser = first(os.Getenv("ADMIN_USER"), cfg.AdminUser, "admin")
	cfg.AdminPassword = first(os.Getenv("ADMIN_PASSWORD"), cfg.AdminPassword)
	cfg.SessionSecret = first(os.Getenv("SESSION_SECRET"), os.Getenv("ADMIN_SECRET"), cfg.SessionSecret, cfg.AdminSecret)
	cfg.R2AccountID = first(os.Getenv("R2_ACCOUNT_ID"), cfg.R2AccountID)
	cfg.R2AccessKeyID = first(os.Getenv("R2_ACCESS_KEY_ID"), cfg.R2AccessKeyID)
	cfg.R2SecretAccessKey = first(os.Getenv("R2_SECRET_ACCESS_KEY"), cfg.R2SecretAccessKey)
	cfg.R2Bucket = first(os.Getenv("R2_BUCKET"), cfg.R2Bucket)
	cfg.R2Endpoint = first(os.Getenv("R2_ENDPOINT"), cfg.R2Endpoint)
	cfg.R2PublicBase = strings.TrimRight(first(os.Getenv("R2_PUBLIC_BASE"), cfg.R2PublicBase), "/")
	cfg.SMTPHost = first(os.Getenv("SMTP_HOST"), cfg.SMTPHost)
	cfg.SMTPPort = first(os.Getenv("SMTP_PORT"), cfg.SMTPPort, "465")
	cfg.SMTPUser = first(os.Getenv("SMTP_USER"), cfg.SMTPUser)
	cfg.SMTPPassword = first(os.Getenv("SMTP_PASSWORD"), cfg.SMTPPassword)
	cfg.SMTPFrom = first(os.Getenv("SMTP_FROM"), cfg.SMTPFrom)
	cfg.AIBailianAPIKey = first(os.Getenv("AI_BAILIAN_API_KEY"), cfg.AIBailianAPIKey)
	cfg.AIBailianChatModels = first(os.Getenv("AI_BAILIAN_CHAT_MODELS"), cfg.AIBailianChatModels)
	if value, ok := os.LookupEnv("DAILY_ARTICLES"); ok {
		cfg.DailyArticles = strings.EqualFold(strings.TrimSpace(value), "true") || value == "1"
	}
	if cfg.R2Endpoint == "" && cfg.R2AccountID != "" {
		cfg.R2Endpoint = "https://" + cfg.R2AccountID + ".r2.cloudflarestorage.com"
	}
	if cfg.SessionSecret == "" {
		return cfg, fmt.Errorf("config.yaml 里要有 sessionSecret")
	}
	return cfg, nil
}

func (c Config) DSN() string {
	dsn := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(c.PostgresUser, c.PostgresPassword),
		Host:   net.JoinHostPort(c.PostgresHost, c.PostgresPort),
		Path:   "/" + c.PostgresDB,
	}
	values := url.Values{"sslmode": {c.PostgresSSLMode}}
	dsn.RawQuery = values.Encode() + "&TimeZone=Asia/Shanghai"
	return dsn.String()
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
