package go_demo

import (
	"context"
	"demo/lock"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"time"
)

func main() {
	backoff := lock.NewExponentialBackoff(
		time.Duration(20)*time.Millisecond,
		time.Duration(1000)*time.Millisecond,
	)
	redisPool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				"redis host",
				redis.DialPassword("redis password"),
			)
		},
	}
	redisLock := lock.NewRedisLock(redisPool, lock.WithBackoff(backoff))
	ctx := context.Background()
	s, err := redisLock.MustSetRetry(ctx, "lock_user")
	if err != nil && err == lock.ErrLockFail {
		fmt.Println(err)
		return
	}

	time.Sleep(20 * time.Second)
	defer func() {
		_ = redisLock.Release(ctx, "lock_user", s)
	}()
	return
}
