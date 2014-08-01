package messages

import (
	"github.com/garyburd/redigo/redis"
)

func newPool(proto, addr, password string) *redis.Pool {
	return redis.NewPool(func() (redis.Conn, error) {

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
