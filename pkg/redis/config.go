package redis

import "github.com/kelseyhightower/envconfig"

// RedisConfig holds the Redis connection configuration.
type RedisConfig struct {
	Addr     string `envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password string `envconfig:"REDIS_PASSWORD" default:""`
	Database int    `envconfig:"REDIS_DATABASE" default:"0"`
}

// LoadRedisConfig loads Redis configuration from environment variables with the given prefix.
func LoadRedisConfig(prefix string) (*RedisConfig, error) {
	cfg := &RedisConfig{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
