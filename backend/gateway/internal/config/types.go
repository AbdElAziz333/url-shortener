package config

import "time"

type AppConfig struct {
	Service ServiceConfig
	Postgres PostgresConfig
	Redis RedisConfig
	JWT JWTConfig
}

type ServiceConfig struct {
	Name string
	Port string
	ShortenerServiceURL string
}

type PostgresConfig struct {
	Host string
	User string
	Password string
	DBName string
}

type RedisConfig struct {
	Addr string
	Password string
}

type JWTConfig struct {
	Secret []byte
	AccessExpiry time.Duration
	RefreshExpiry time.Duration
}