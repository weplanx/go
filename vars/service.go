package vars

import (
	"context"
	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"time"
)

type Service struct {
	Db    *mongo.Database
	Redis *redis.Client
}

// Refresh 刷新变量缓存
func (x *Service) Refresh(ctx context.Context) (err error) {
	var exists int64
	if exists, err = x.Redis.Exists(ctx, Key).Result(); err != nil {
		return
	}
	if exists != 0 {
		return
	}
	var cursor *mongo.Cursor
	if cursor, err = x.Db.Collection("vars").
		Find(ctx, bson.M{}); err != nil {
		return
	}
	var vars []Var
	if err = cursor.All(ctx, &vars); err != nil {
		return
	}
	values := make(map[string]interface{})
	for _, data := range vars {
		switch value := data.Value.(type) {
		case primitive.A:
			b, _ := jsoniter.Marshal(value)
			values[data.Key] = b
			break
		case primitive.M:
			b, _ := jsoniter.Marshal(value)
			values[data.Key] = b
			break
		default:
			values[data.Key] = value
		}
	}
	return x.Redis.HSet(ctx, Key, values).Err()
}

func (x *Service) Get(ctx context.Context, key string) (value string, err error) {
	if err = x.Refresh(ctx); err != nil {
		return
	}
	return x.Redis.HGet(ctx, Key, key).Result()
}

func (x *Service) Set(ctx context.Context, key string, value interface{}) (err error) {
	var exists int64
	if exists, err = x.Db.Collection("vars").
		CountDocuments(ctx, bson.M{"key": key}); err != nil {
		return
	}
	doc := &Var{Key: key, Value: value}
	if exists == 0 {
		if _, err = x.Db.Collection("vars").
			InsertOne(ctx, doc); err != nil {
			return
		}
	} else {
		if _, err = x.Db.Collection("vars").
			ReplaceOne(ctx, bson.M{"key": key}, doc); err != nil {
			return
		}
	}
	if err = x.Redis.Del(ctx, Key).Err(); err != nil {
		return
	}
	return
}

func (x *Service) toMap(ctx context.Context, keys []string) (result map[string]interface{}, err error) {
	if err = x.Refresh(ctx); err != nil {
		return
	}
	var data []interface{}
	if data, err = x.Redis.HMGet(ctx, Key, keys...).Result(); err != nil {
		return
	}
	result = make(map[string]interface{})
	for k, v := range keys {
		result[v] = data[k]
	}
	return
}

func (x *Service) ParseDuration(ctx context.Context, key string) (_ time.Duration, err error) {
	var value string
	if value, err = x.Get(ctx, key); err != nil {
		return
	}
	return time.ParseDuration(value)
}

func (x *Service) Atoi(ctx context.Context, key string) (_ int, err error) {
	var value string
	if value, err = x.Get(ctx, key); err != nil {
		return
	}
	return strconv.Atoi(value)
}

func (x *Service) Unmarshal(ctx context.Context, key string, v interface{}) (err error) {
	var value string
	if value, err = x.Get(ctx, key); err != nil {
		return
	}
	return jsoniter.Unmarshal([]byte(value), v)
}

func (x *Service) GetUserSessionExpire(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "user_session_expire")
}

func (x *Service) GetUserLoginFailedTimes(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "user_login_failed_times")
}

func (x *Service) GetUserLockTime(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "user_lock_time")
}

func (x *Service) GetIpLoginFailedTimes(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "ip_login_failed_times")
}

func (x *Service) GetIpWhitelist(ctx context.Context) (result []string, err error) {
	if err = x.Unmarshal(ctx, "ip_whitelist", &result); err != nil {
		return
	}
	return
}

func (x *Service) GetIpBlacklist(ctx context.Context) (result []string, err error) {
	if err = x.Unmarshal(ctx, "ip_blacklist", &result); err != nil {
		return
	}
	return
}

func (x *Service) GetPasswordStrength(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "password_strength")
}

func (x *Service) GetPasswordExpire(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "password_expire")
}

func (x *Service) GetCloudPlatform(ctx context.Context) (string, error) {
	return x.Get(ctx, "cloud_platform")
}

func (x *Service) GetTencentSecretId(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_secret_id")
}

func (x *Service) GetTencentSecretKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_secret_key")
}

func (x *Service) GetTencentCosBucket(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_cos_bucket")
}

func (x *Service) GetTencentCosRegion(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_cos_region")
}

func (x *Service) GetTencentCosExpired(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "tencent_cos_expired")
}

func (x *Service) GetTencentCosLimit(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "tencent_cos_limit")
}

func (x *Service) GetOfficePlatform(ctx context.Context) (string, error) {
	return x.Get(ctx, "office_platform")
}

func (x *Service) GetFeishuAppId(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_app_id")
}

func (x *Service) GetFeishuAppSecret(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_app_secret")
}

func (x *Service) GetFeishuEncryptKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_encrypt_key")
}

func (x *Service) GetFeishuVerificationToken(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_verification_token")
}

func (x *Service) GetRedirectUrl(ctx context.Context) (string, error) {
	return x.Get(ctx, "redirect_url")
}

func (x *Service) GetEmailHost(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_host")
}

func (x *Service) GetEmailPort(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_port")
}

func (x *Service) GetEmailUsername(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_username")
}

func (x *Service) GetEmailPassword(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_password")
}

func (x *Service) GetOpenapiUrl(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_url")
}

func (x *Service) GetOpenapiKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_key")
}

func (x *Service) GetOpenapiSecret(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_secret")
}
