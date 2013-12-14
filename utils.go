package messages

import (
	"github.com/garyburd/redigo/redis"
	"os"
)

func newPool() *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {
		var (
			proto    = defaultEnv("REDIS_PROTO", "tcp")
			addr     = defaultEnv("REDIS_ADDR", "127.0.0.1:6379")
			password = defaultEnv("REDIS_PASSWORD", "")
		)

		c, err := redis.Dial(proto, addr)
		if err != nil {
			return nil, err
		}

		if password != "" {
			if _, err := c.Do("AUTH", password); err != nil {
				return nil, err
			}
		}
		return c, nil
	}, DefaultPoolSize)
}

func argsToMap(args [][]byte) map[string][]byte {
	result := make(map[string][]byte, len(args)/2)
	for i := 0; i < len(args); i++ {
		key := string(args[i])
		i++
		result[key] = args[i]
	}
	return result
}

// Get a value from the machines environment or use the
// default value
func defaultEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}
	return value
}
