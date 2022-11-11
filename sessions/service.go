package sessions

import (
	"context"
	"fmt"
	"strings"
)

type Service struct {
	*Sessions
}

func (x *Service) Key(name string) string {
	return fmt.Sprintf(`%s:sessions:%s`, x.Namespace, name)
}

// Lists 列出所有会话用户 ID
func (x *Service) Lists(ctx context.Context) (data []string, err error) {
	var cursor uint64
	for cursor != 0 {
		var keys []string
		if keys, cursor, err = x.Redis.
			Scan(ctx, cursor, x.Key("*"), 1000).
			Result(); err != nil {
			return
		}
		names := make([]string, len(keys))
		for k, v := range keys {
			names[k] = strings.Replace(v, x.Key(""), "", -1)
		}
		data = append(data, names...)
	}
	return
}

// Verify 验证会话一致性
func (x *Service) Verify(ctx context.Context, name string, jti string) (result bool, err error) {
	var value string
	if value, err = x.Redis.
		Get(ctx, x.Key(name)).
		Result(); err != nil {
		return
	}
	return value == jti, nil
}

// Set 设置会话
func (x *Service) Set(ctx context.Context, name string, jti string) error {
	return x.Redis.
		Set(ctx, x.Key(name), jti, x.Values.SessionTTL).
		Err()
}

// Renew 续约会话
func (x *Service) Renew(ctx context.Context, userId string) error {
	return x.Redis.
		Expire(ctx, x.Key(userId), x.Values.SessionTTL).
		Err()
}

// Remove 移除会话
func (x *Service) Remove(ctx context.Context, name string) error {
	return x.Redis.
		Del(ctx, x.Key(name)).
		Err()
}

// Clear 清除所有会话
func (x *Service) Clear(ctx context.Context) (err error) {
	pipe := x.Redis.TxPipeline()
	var cursor uint64
	for cursor != 0 {
		var keys []string
		if keys, cursor, err = x.Redis.
			Scan(ctx, cursor, x.Key("*"), 1000).
			Result(); err != nil {
			return
		}
		pipe.Del(ctx, keys...)
	}
	if _, err = pipe.Exec(ctx); err != nil {
		return
	}
	return
}