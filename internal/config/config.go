package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env          string   `yaml:"env" env-default:"local"`
	AllowOrigins []string `yaml:"allow_origins" env-default:"*"`
	HttpServer   `yaml:"http_server"`
	DBSettings   `yaml:"db"`
	RedisServer  `yaml:"redis"`
	Token        `yaml:"token"`
	SMTP         `yaml:"smtp"`
	MinIO        `yaml:"minio"`
	Log          `yaml:"log"`
}

// Log controls the structured logger built by pkg/logger.
type Log struct {
	// Level is one of debug|info|warn|error.
	Level string `yaml:"level" env-default:"info"`
	// AddCaller adds the source file:line of each log call.
	AddCaller bool `yaml:"add_caller" env-default:"true"`
	// Stacktrace is the minimum level a stacktrace is attached at.
	Stacktrace string `yaml:"stacktrace" env-default:"error"`
}

type HttpServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type DBSettings struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     int    `yaml:"port" env-default:"5436"`
	Username string `yaml:"username" env-default:"postgres"`
	Password string
	DBName   string `yaml:"dbname" env-default:"postgres"`
	SSLMode  string `yaml:"sslmode" env-default:"disable"`
}

type RedisServer struct {
	Address  string   `yaml:"address" env-default:"localhost:6380"`
	Username string   `yaml:"username" env-required:"true"`
	DB       int      `yaml:"db" env-default:"0"`
	Password string   `yaml:"password" env-required:"true"`
	TTL      RedisTTL `yaml:"ttl"`
}

type RedisTTL struct {
	UserProfile      time.Duration `yaml:"user_profile" env-default:"1h"`
	BlacklistAccess  time.Duration `yaml:"blacklist_access" env-default:"15m"`
	BlacklistRefresh time.Duration `yaml:"blacklist_refresh" env-default:"168h"`
	AuthTimeout      time.Duration `yaml:"auth_timeout" env-default:"5m"`
}

type Token struct {
	PasetoKeyHex string        `yaml:"paseto_key_hex"`
	Cookie       Cookie        `yaml:"cookie"`
	AccessTTL    time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL   time.Duration `yaml:"refresh_ttl" env-default:"168h"`
}

type Cookie struct {
	Name     string `yaml:"name" env-default:"refresh_token"`
	Domain   string `yaml:"domain" env-default:""`
	Path     string `yaml:"path" env-default:"/"`
	Secure   bool   `yaml:"secure" env-default:"false"`
	SameSite string `yaml:"same_site" env-default:"lax"`
}

type SMTP struct {
	Host      string `yaml:"host" env-default:"smtp.yandex.ru"`
	Port      int    `yaml:"port" env-default:"587"`
	Username  string
	Password  string
	From      string
	WorkEmail string
}

type MinIO struct {
	Endpoint  string `yaml:"endpoint" env-default:"localhost:9000"`
	AccessKey string
	SecretKey string
	Bucket    string `yaml:"bucket" env-default:"ipbuild-unet-bucket"`
	UseSSL    bool   `yaml:"use_ssl" env-default:"false"`
}

// Load reads public configuration from a YAML file and secrets from an env file.
//
//	configFile — path to the YAML file, e.g. "config/local.yaml"
//	envFile    — path to the secrets .env file, e.g. ".env"
func Load(configFile, envFile string) (*Config, error) {
	// ── YAML: public / non-sensitive settings ────────────────────────────────
	yv := viper.New()
	yv.SetConfigFile(configFile)
	yv.SetConfigType("yaml")
	yv.SetDefault("env", "local")
	yv.SetDefault("allow_origins", []string{"*"})
	yv.SetDefault("http_server.address", "localhost:8080")
	yv.SetDefault("http_server.timeout", 4*time.Second)
	yv.SetDefault("http_server.idle_timeout", time.Minute)
	yv.SetDefault("db.host", "localhost")
	yv.SetDefault("db.port", "5432")
	yv.SetDefault("db.username", "postgres")
	yv.SetDefault("db.dbname", "postgres")
	yv.SetDefault("db.sslmode", "disable")
	yv.SetDefault("redis.address", "localhost:6380")
	yv.SetDefault("redis.db", 0)
	yv.SetDefault("redis.ttl.user_profile", 1*time.Hour)
	yv.SetDefault("redis.ttl.blacklist_access", 15*time.Minute)
	yv.SetDefault("redis.ttl.blacklist_refresh", 168*time.Hour)
	yv.SetDefault("redis.ttl.auth_timeout", 5*time.Minute)
	yv.SetDefault("token.access_ttl", 15*time.Minute)
	yv.SetDefault("token.refresh_ttl", 168*time.Hour)
	yv.SetDefault("smtp.host", "smtp.yandex.ru")
	yv.SetDefault("smtp.port", 587)
	yv.SetDefault("minio.endpoint", "localhost:9000")
	yv.SetDefault("minio.use_ssl", false)
	yv.SetDefault("minio.bucket", "ipbuild-unet-bucket")
	yv.SetDefault("log.level", "info")
	yv.SetDefault("log.add_caller", true)
	yv.SetDefault("log.stacktrace", "error")
	yv.SetDefault("token.cookie.name", "refresh_token")
	yv.SetDefault("token.cookie.domain", "")
	yv.SetDefault("token.cookie.path", "/")
	yv.SetDefault("token.cookie.secure", false)
	yv.SetDefault("token.cookie.same_site", "lax")

	if err := yv.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read yaml %q: %w", configFile, err)
	}

	// ── .env: secrets ─────────────────────────────────────────────────────────
	ev := viper.New()
	ev.SetConfigFile(envFile)
	ev.SetConfigType("env")
	if err := ev.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: read env %q: %w", envFile, err)
	}

	// ── Assemble ──────────────────────────────────────────────────────────────
	cfg := &Config{}
	cfg.Env = yv.GetString("env")
	cfg.AllowOrigins = yv.GetStringSlice("allow_origins")

	cfg.HttpServer.Address = yv.GetString("http_server.address")
	cfg.HttpServer.Timeout = yv.GetDuration("http_server.timeout")
	cfg.HttpServer.IdleTimeout = yv.GetDuration("http_server.idle_timeout")

	cfg.DBSettings.Host = yv.GetString("db.host")
	cfg.DBSettings.Port = yv.GetInt("db.port")
	cfg.DBSettings.Username = yv.GetString("db.username")
	cfg.DBSettings.DBName = yv.GetString("db.dbname")
	cfg.DBSettings.Password = ev.GetString("DB_PASSWORD")
	cfg.DBSettings.SSLMode = yv.GetString("db.sslmode")

	cfg.RedisServer.Address = yv.GetString("redis.address")
	cfg.RedisServer.Username = yv.GetString("redis.username")
	cfg.RedisServer.DB = yv.GetInt("redis.db")
	cfg.RedisServer.Password = ev.GetString("REDIS_PASSWORD")

	cfg.RedisServer.TTL.UserProfile = yv.GetDuration("redis.ttl.user_profile")
	cfg.RedisServer.TTL.BlacklistAccess = yv.GetDuration("redis.ttl.blacklist_access")
	cfg.RedisServer.TTL.BlacklistRefresh = yv.GetDuration("redis.ttl.blacklist_refresh")
	cfg.RedisServer.TTL.AuthTimeout = yv.GetDuration("redis.ttl.auth_timeout")

	cfg.Token.PasetoKeyHex = ev.GetString("PASETO_KEY")
	cfg.Token.AccessTTL = yv.GetDuration("token.access_ttl")
	cfg.Token.RefreshTTL = yv.GetDuration("token.refresh_ttl")
	cfg.Token.Cookie.Name = yv.GetString("token.cookie.name")
	cfg.Token.Cookie.Domain = yv.GetString("token.cookie.domain")
	cfg.Token.Cookie.Path = yv.GetString("token.cookie.path")
	cfg.Token.Cookie.Secure = yv.GetBool("token.cookie.secure")
	cfg.Token.Cookie.SameSite = yv.GetString("token.cookie.same_site")

	cfg.SMTP.Host = yv.GetString("smtp.host")
	cfg.SMTP.Port = yv.GetInt("smtp.port")
	cfg.SMTP.Username = ev.GetString("SMTP_USERNAME")
	cfg.SMTP.Password = ev.GetString("SMTP_PASSWORD")
	cfg.SMTP.From = ev.GetString("SMTP_FROM")
	cfg.SMTP.WorkEmail = ev.GetString("WORK_EMAIL")

	cfg.MinIO.Endpoint = yv.GetString("minio.endpoint")
	cfg.MinIO.AccessKey = ev.GetString("MINIO_ACCESS_KEY")
	cfg.MinIO.SecretKey = ev.GetString("MINIO_SECRET_KEY")
	cfg.MinIO.Bucket = yv.GetString("minio.bucket")
	cfg.MinIO.UseSSL = yv.GetBool("minio.use_ssl")

	cfg.Log.Level = yv.GetString("log.level")
	cfg.Log.AddCaller = yv.GetBool("log.add_caller")
	cfg.Log.Stacktrace = yv.GetString("log.stacktrace")

	return cfg, nil
}
