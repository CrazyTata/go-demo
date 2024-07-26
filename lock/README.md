the lock use exponential backoff

## Usage

```go
func Lock() {
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
                l.svcCtx.Config.RedisCache.Host,
                redis.DialPassword(l.svcCtx.Config.RedisCache.Pass),
            )
        },
    }
    redisLock := lock.NewRedisLock(redisPool, lock.WithBackoff(backoff))
    ctx := context.Background()
    var userLockInstance lock.LockInstanceIFace = lock.NewLockInstance(redisLock)
    if err := userLockInstance.MustSetRetry(ctx, "lock_user"); err != nil {
        if err == lock.ErrLockFail {
            l.Logger.Errorf("lock.ErrLockFail error:%+v", err)
            return
        }
        return
    }
    time.Sleep(20 * time.Second)
    defer func() {
        _ = userLockInstance.Release(ctx)
    }()
    return
}

```

