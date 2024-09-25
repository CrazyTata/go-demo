package lock

import (
	"context"
	"errors"
)

var (
	ErrLockExists   = errors.New("已存在锁")
	ErrLockKeyEmpty = errors.New("锁KEY不能为空")
)

type LockInstanceIFace interface {
	MustSet(ctx context.Context, key string) error
	MustSetRetry(ctx context.Context, key string) error
	Release(context.Context) error
}

type LockInstance struct {
	key    string
	ranVal string
	RedisLockIFace
}

func NewLockInstance(redisLockIFace RedisLockIFace) *LockInstance {
	return &LockInstance{RedisLockIFace: redisLockIFace}
}

func (p *LockInstance) validKey(key string) error {
	if key == "" {
		return ErrLockKeyEmpty
	}
	if p.key != "" {
		return ErrLockExists
	}
	return nil
}

func (p *LockInstance) MustSet(ctx context.Context, key string) error {

	if err := p.validKey(key); err != nil {
		return err
	}

	r, err := p.RedisLockIFace.MustSet(ctx, key)
	if err != nil {
		return err
	}

	p.key = key
	p.ranVal = r

	return nil
}

func (p *LockInstance) MustSetRetry(ctx context.Context, key string) error {

	if err := p.validKey(key); err != nil {
		return err
	}

	r, err := p.RedisLockIFace.MustSetRetry(ctx, key)
	if err != nil {
		return err
	}

	p.key = key
	p.ranVal = r

	return nil
}

func (p *LockInstance) IsEmpty() bool {
	return p.key == "" || p.ranVal == ""
}

func (p *LockInstance) Release(ctx context.Context) error {

	if p.IsEmpty() {
		return nil
	}

	return p.RedisLockIFace.Release(ctx, p.key, p.ranVal)
}
