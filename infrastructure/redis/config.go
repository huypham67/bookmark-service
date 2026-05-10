package redis

// RedisConfig holds the configuration parameters for connecting to a Redis server.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	Database int
}
