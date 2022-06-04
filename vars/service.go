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

// GetUserSessionExpire 用户无操作的最大时长，超出将结束会话。
func (x *Service) GetUserSessionExpire(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "user_session_expire")
}

// GetUserLoginFailedTimes 用户有限时间（等同锁定时间）内连续登录失败的次数，超出锁定帐号。
func (x *Service) GetUserLoginFailedTimes(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "user_login_failed_times")
}

// GetUserLockTime 锁定账户时间。
func (x *Service) GetUserLockTime(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "user_lock_time")
}

// GetIpLoginFailedTimes IP 连续登录失败后的最大次数（无白名单时启用），锁定 IP。
func (x *Service) GetIpLoginFailedTimes(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "ip_login_failed_times")
}

// GetIpWhitelist IP 白名单。
func (x *Service) GetIpWhitelist(ctx context.Context) (result []string, err error) {
	if err = x.Unmarshal(ctx, "ip_whitelist", &result); err != nil {
		return
	}
	return
}

// GetIpBlacklist IP 黑名单。
func (x *Service) GetIpBlacklist(ctx context.Context) (result []string, err error) {
	if err = x.Unmarshal(ctx, "ip_blacklist", &result); err != nil {
		return
	}
	return
}

// GetPasswordStrength 密码强度。
// 0：无限制；1：需要大小写字母；2：需要大小写字母、数字；3：需要大小写字母、数字、特殊字符。
func (x *Service) GetPasswordStrength(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "password_strength")
}

// GetPasswordExpire 密码有效期（天）。
// 密码过期后强制要求修改密码，0：永久有效
func (x *Service) GetPasswordExpire(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "password_expire")
}

// GetCloudPlatform 云厂商。
// tencent：腾讯云；
func (x *Service) GetCloudPlatform(ctx context.Context) (string, error) {
	return x.Get(ctx, "cloud_platform")
}

// GetTencentSecretId 腾讯云 API 密钥 Id，建议用子账号分配需要的权限
func (x *Service) GetTencentSecretId(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_secret_id")
}

// GetTencentSecretKey 腾讯云 API 密钥 Key
func (x *Service) GetTencentSecretKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_secret_key")
}

// GetTencentCosBucket 腾讯云 COS 对象存储 Bucket（存储桶名称）
func (x *Service) GetTencentCosBucket(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_cos_bucket")
}

// GetTencentCosRegion 腾讯云 COS 对象存储所属地域，例如：ap-guangzhou
func (x *Service) GetTencentCosRegion(ctx context.Context) (string, error) {
	return x.Get(ctx, "tencent_cos_region")
}

// GetTencentCosExpired 腾讯云 COS 对象存储预签名有效期，单位：秒
func (x *Service) GetTencentCosExpired(ctx context.Context) (time.Duration, error) {
	return x.ParseDuration(ctx, "tencent_cos_expired")
}

// GetTencentCosLimit 腾讯云 COS 对象存储上传大小限制，单位：KB
func (x *Service) GetTencentCosLimit(ctx context.Context) (int, error) {
	return x.Atoi(ctx, "tencent_cos_limit")
}

// GetOfficePlatform 办公平台。
// feishu：飞书；
func (x *Service) GetOfficePlatform(ctx context.Context) (string, error) {
	return x.Get(ctx, "office_platform")
}

// GetFeishuAppId 飞书应用 ID
func (x *Service) GetFeishuAppId(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_app_id")
}

// GetFeishuAppSecret 飞书应用密钥
func (x *Service) GetFeishuAppSecret(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_app_secret")
}

// GetFeishuEncryptKey 飞书事件订阅安全校验数据密钥
func (x *Service) GetFeishuEncryptKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_encrypt_key")
}

// GetFeishuVerificationToken 飞书事件订阅验证令牌
func (x *Service) GetFeishuVerificationToken(ctx context.Context) (string, error) {
	return x.Get(ctx, "feishu_verification_token")
}

// GetRedirectUrl 第三方免登授权码跳转地址
func (x *Service) GetRedirectUrl(ctx context.Context) (string, error) {
	return x.Get(ctx, "redirect_url")
}

// GetEmailHost 公共电子邮件服务 SMTP 地址
func (x *Service) GetEmailHost(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_host")
}

// GetEmailPort SMTP 端口号（SSL）
func (x *Service) GetEmailPort(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_port")
}

// GetEmailUsername 公共邮箱用户，例如：support@example.com
func (x *Service) GetEmailUsername(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_username")
}

// GetEmailPassword 公共邮箱用户密码
func (x *Service) GetEmailPassword(ctx context.Context) (string, error) {
	return x.Get(ctx, "email_password")
}

// GetOpenapiUrl 开放服务地址
func (x *Service) GetOpenapiUrl(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_url")
}

// GetOpenapiKey 开放服务应用认证 Key
// API 网关应用认证方式 https://cloud.tencent.com/document/product/628/55088
func (x *Service) GetOpenapiKey(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_key")
}

// GetOpenapiSecret 开放服务应用认证密钥
func (x *Service) GetOpenapiSecret(ctx context.Context) (string, error) {
	return x.Get(ctx, "openapi_secret")
}
