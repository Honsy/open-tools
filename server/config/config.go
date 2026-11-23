package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MySQLHost     string `yaml:"mysqlHost"`
	MySQLPort     string `yaml:"mysqlPort"`
	MySQLUser     string `yaml:"mysqlUser"`
	MySQLPassword string `yaml:"mysqlPassword"`
	MySQLDB       string `yaml:"mysqlDB"`
	AdminUser     string `yaml:"adminUser"`
	AdminPassword string `yaml:"adminPassword"`
	SessionSecret string `yaml:"sessionSecret"`
	AdminSecret   string `yaml:"adminSecret"`

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
		MySQLHost: "127.0.0.1",
		MySQLPort: "3306",
		MySQLUser: "root",
		MySQLDB:   "opentools",
		AdminUser: "admin",
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
	cfg.MySQLHost = first(os.Getenv("YOUGO_MYSQL_HOST"), os.Getenv("MYSQL_HOST"), cfg.MySQLHost)
	cfg.MySQLPort = first(os.Getenv("YOUGO_MYSQL_PORT"), os.Getenv("MYSQL_PORT"), cfg.MySQLPort)
	cfg.MySQLUser = first(os.Getenv("YOUGO_MYSQL_USER"), os.Getenv("MYSQL_USER"), cfg.MySQLUser)
	if v, ok := os.LookupEnv("YOUGO_MYSQL_PASSWORD"); ok {
		cfg.MySQLPassword = v
	} else if v, ok := os.LookupEnv("MYSQL_PASSWORD"); ok {
		cfg.MySQLPassword = v
	}
	cfg.MySQLDB = first(os.Getenv("MYSQL_DB"), cfg.MySQLDB)
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

func (c Config) ServerDSN() string {
	return c.dsn("")
}

func (c Config) DSN() string {
	return c.dsn(c.MySQLDB)
}

func (c Config) dsn(name string) string {
	mc := mysql.Config{
		User:                 c.MySQLUser,
		Passwd:               c.MySQLPassword,
		Net:                  "tcp",
		Addr:                 c.MySQLHost + ":" + c.MySQLPort,
		DBName:               name,
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
		Params:               map[string]string{"charset": "utf8mb4"},
	}
	return mc.FormatDSN()
}

func first(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
