package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	HttpServer  `yaml:"http_server"`
	DBSettings  `yaml:"db"`
	RedisServer `yaml:"redis"`
	Token       `yaml:"token"`
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
	UserProfile time.Duration `yaml:"user_profile" env-default:"15m"`
	Blacklist   time.Duration `yaml:"blacklist" env-default:"5m"`
	AuthTimeout time.Duration `yaml:"auth_timeout" env-default:"5m"`
}

type Token struct {
	PasetoKeyHex string        `yaml:"paseto_key_hex"`
	AccessTTL    time.Duration `yaml:"access_ttl" env-default:"15m"`
	RefreshTTL   time.Duration `yaml:"refresh_ttl" env-default:"168h"`
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
	yv.SetDefault("redis.ttl.user_profile", 15*time.Minute)
	yv.SetDefault("redis.ttl.blacklist", 5*time.Minute)
	yv.SetDefault("redis.ttl.auth_timeout", 5*time.Minute)
	yv.SetDefault("token.access_ttl", 15*time.Minute)
	yv.SetDefault("token.refresh_ttl", 168*time.Hour)

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
	cfg.RedisServer.TTL.Blacklist = yv.GetDuration("redis.ttl.blacklist")
	cfg.RedisServer.TTL.AuthTimeout = yv.GetDuration("redis.ttl.auth_timeout")

	cfg.Token.PasetoKeyHex = ev.GetString("PASETO_KEY")
	cfg.Token.AccessTTL = yv.GetDuration("token.access_ttl")
	cfg.Token.RefreshTTL = yv.GetDuration("token.refresh_ttl")

	return cfg, nil
}
