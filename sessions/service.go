package sessions

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/weplanx/go/values"
	"strings"
)

type Service struct {
	RDb    *redis.Client
	Values *values.DynamicValues
}

func (x *Service) Key(name string) string {
	return fmt.Sprintf(`sessions:%s`, name)
}

type ScanFn func(key string)

func (x *Service) Scan(ctx context.Context, fn ScanFn) {
	iter := x.RDb.Scan(ctx, 0, x.Key("*"), 0).Iterator()
	for iter.Next(ctx) {
		fn(iter.Val())
	}
}

func (x *Service) Lists(ctx context.Context) (data []string) {
	data = make([]string, 0)
	x.Scan(ctx, func(key string) {
		v := strings.Replace(key, x.Key(""), "", -1)
		data = append(data, v)
	})
	return
}

func (x *Service) Verify(ctx context.Context, name string, jti string) bool {
	result := x.RDb.Get(ctx, x.Key(name)).Val()
	return result == jti
}

func (x *Service) Set(ctx context.Context, name string, jti string) string {
	return x.RDb.Set(ctx, x.Key(name), jti, x.Values.SessionTTL).Val()
}

func (x *Service) Renew(ctx context.Context, userId string) bool {
	return x.RDb.Expire(ctx, x.Key(userId), x.Values.SessionTTL).Val()
}

func (x *Service) Remove(ctx context.Context, name string) int64 {
	return x.RDb.Del(ctx, x.Key(name)).Val()
}

func (x *Service) Clear(ctx context.Context) int64 {
	var matchd []string
	x.Scan(ctx, func(key string) {
		matchd = append(matchd, key)
	})
	if len(matchd) != 0 {
		return x.RDb.Del(ctx, matchd...).Val()
	}
	return 0
}
