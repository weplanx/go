package session

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

var Inconsistent = errors.New("verification is inconsistent")

type Session struct {
	Redis *redis.Client
	Exp   time.Duration
}

func New(r *redis.Client, exp time.Duration) *Session {
	return &Session{
		Redis: r,
		Exp:   exp,
	}
}

func (x *Session) Update(jti string, uid string) error {
	return x.Redis.Set(context.Background(), "session:"+uid, jti, x.Exp).Err()
}

func (x *Session) Check(jti string, uid string) error {
	result, err := x.Redis.Get(context.Background(), "session:"+uid).Result()
	if err != nil {
		return err
	}
	if result != jti {
		return Inconsistent
	}
	return nil
}

func (x *Session) Destory(uid string) error {
	return x.Redis.Del(context.Background(), "session:"+uid).Err()
}
