package lock

import (
	"context"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	// 防止孤儿lock没release
	// 目前expire过期时间的敏感度是考虑为一致的敏感度
	defaultExpireSecond uint32 = 30
)

var (
	ErrLockSet     = errors.New("lock set err")
	ErrLockRelease = errors.New("lock release err")
	ErrLockFail    = errors.New("lock fail")
)

// RedisLockIFace 在common redis上封一层浅封装
// 将redis pool 与expire second作为redis lock已知数据
type RedisLockIFace interface {
	MustSet(ctx context.Context, k string) (string, error)
	MustSetRetry(ctx context.Context, k string) (string, error) // 必须设置成功并有重试机制
	Release(ctx context.Context, k string, randVal string) error
}

// RedisLock nil的实现默认为true
type RedisLock struct {
	redisPool    *redis.Pool
	expireSecond uint32
	backoff      Backoff
}

// An Option configures a RedisLock.
type Option interface {
	apply(*RedisLock)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*RedisLock)

func (f optionFunc) apply(log *RedisLock) {
	f(log)
}

// WithBackoff backoff set
func WithBackoff(b Backoff) Option {
	return optionFunc(func(r *RedisLock) {
		r.backoff = b
	})
}

func NewRedisLock(redisPool *redis.Pool, opts ...Option) *RedisLock {

	r := &RedisLock{
		redisPool:    redisPool,
		expireSecond: defaultExpireSecond,
		backoff:      NewExponentialBackoff(30*time.Millisecond, 500*time.Millisecond), // default backoff
	}

	for _, opt := range opts {
		opt.apply(r)
	}

	return r
}

func (r *RedisLock) Set(ctx context.Context, key string) (bool, string, error) {

	if r == nil {
		return true, "", nil
	}

	isLock, randVal, err := SetWithContext(ctx, r.redisPool, key, r.expireSecond)
	if err != nil {
		return isLock, randVal, ErrLockSet
	}

	return isLock, randVal, err
}

// MustSetRetry 必须设置成功并带有重试功能
func (r *RedisLock) MustSetRetry(ctx context.Context, key string) (string, error) {

	op := func() (string, error) {
		return r.MustSet(ctx, key)
	}

	notifyFunc := func(err error) {
		if err == ErrLockFail {
			logx.WithContext(ctx).Errorf("RedisLock.MustSetRetry redis must set err: %v", err)
		} else {
			logx.WithContext(ctx).Errorf("RedisLock.MustSetRetry redis must set err: %v", err)
		}
	}

	return mustSetRetryNotify(op, r.backoff, notifyFunc)
}

func (r *RedisLock) MustSet(ctx context.Context, key string) (string, error) {

	isLock, randVal, err := r.Set(ctx, key)
	if err != nil {
		return "", err
	}

	if !isLock {
		return "", ErrLockFail
	}

	return randVal, nil
}

func (r *RedisLock) Release(ctx context.Context, key string, randVal string) error {

	if r == nil {
		logx.WithContext(ctx).Infof("that the implementation of redis lock is nil")
		return nil
	}

	err := ReleaseWithContext(ctx, r.redisPool, key, randVal)
	if err != nil {
		logx.WithContext(ctx).Errorf("s.RedisLock.ReleaseWithContext fail, err: %v", err)
		return ErrLockRelease
	}

	return nil
}

func SetWithContext(ctx context.Context, redisPool *redis.Pool, key string, expireSecond uint32) (bool, string, error) {
	if expireSecond == 0 {
		return false, "", fmt.Errorf("expireSecond参数必须大于0")
	}

	conn, _ := redisPool.GetContext(ctx)
	defer conn.Close()

	randVal := time.Now().Format("2006-01-02 15:04:05.000")
	reply, err := conn.Do("SET", key, randVal, "NX", "PX", expireSecond*1000)
	if err != nil {
		return false, "", err
	}
	if reply == nil {
		return false, "", nil
	}

	return true, randVal, nil
}

func ReleaseWithContext(ctx context.Context, redisPool *redis.Pool, key string, randVal string) error {
	conn, _ := redisPool.GetContext(ctx)
	defer conn.Close()

	luaScript := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end;
	`
	script := redis.NewScript(1, luaScript)
	_, err := script.Do(conn, key, randVal)

	return err
}
