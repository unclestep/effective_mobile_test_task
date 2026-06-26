package config

import (
	"fmt"
	"net/url"

	"em"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Logger   LoggerConfig   `yaml:"logger"`
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres-database"`
}

type LoggerConfig struct {
	Development bool `yaml:"development" env:"LOGGER_DEV"`
}

type ServerConfig struct {
	Addr string `yaml:"addr" env:"SERVER_ADDR"`
}

type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT"`
	User     string `yaml:"user" env:"POSTGRES_USER"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"name" env:"POSTGRES_DB"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSL"`

	MaxConns          int32    `yaml:"max_conns" env:"POSTGRES_MAX_CONNS"`
	MinConns          int32    `yaml:"min_conns" env:"POSTGRES_MIN_CONNS"`
	MaxConnLifetime   Duration `yaml:"max_conn_lifetime" env:"POSTGRES_MAX_CONN_LIFETIME"`
	MaxConnIdleTime   Duration `yaml:"max_conn_idle_time" env:"POSTGRES_MAX_CONN_IDLE_TIME"`
	HealthCheckPeriod Duration `yaml:"health_check_period" env:"POSTGRES_HEALTH_CHECK_PERIOD"`
	ConnectTimeout    Duration `yaml:"connect_timeout" env:"POSTGRES_CONNECT_TIMEOUT"`
	PoolInitTimeout   Duration `yaml:"pool_init_timeout" env:"POSTGRES_POOL_INIT_TIMEOUT"`
}

func (p PostgresConfig) DSN() string {
	sslMode := p.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.User, p.Password),
		Host:     fmt.Sprintf("%s:%d", p.Host, p.Port),
		Path:     "/" + p.DBName,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}

	return u.String()
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	f, err := em.ConfigFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load config: open embedded file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := cleanenv.ParseYAML(f, &cfg); err != nil {
		return nil, fmt.Errorf("load config: parse yaml: %w", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("load config: read env: %w", err)
	}

	return &cfg, nil
}
