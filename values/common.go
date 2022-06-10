package values

import (
	"github.com/google/wire"
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"time"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type Values struct {
	// 用户无操作的最大时长，超出将结束会话
	UserSessionExpire time.Duration `json:"user_session_expire"`

	// 用户有限时间（等同锁定时间）内连续登录失败的次数，超出锁定帐号
	UserLoginFailedTimes int64 `json:"user_login_failed_times"`

	// 锁定账户时间
	UserLockTime time.Duration `json:"user_lock_time"`

	// IP 连续登录失败后的最大次数（无白名单时启用），锁定 IP
	IpLoginFailedTimes int64 `json:"ip_login_failed_times"`

	// IP 白名单
	IpWhitelist []string `json:"ip_whitelist"`

	// IP 黑名单
	IpBlacklist []string `json:"ip_blacklist"`

	// 密码强度
	// 0：无限制；1：需要大小写字母；2：需要大小写字母、数字；3：需要大小写字母、数字、特殊字符
	PasswordStrength int64 `json:"password_strength"`

	// 密码有效期（天）
	// 密码过期后强制要求修改密码，0：永久有效
	PasswordExpire int64 `json:"password_expire"`

	// 云厂商
	// 腾讯云
	CloudPlatform string `json:"cloud_platform"`

	// 腾讯云 API 密钥 Id，建议用子账号分配需要的权限
	TencentSecretId string `json:"tencent_secret_id"`

	// 腾讯云 API 密钥 Key
	TencentSecretKey string `json:"tencent_secret_key"`

	// 腾讯云 COS 对象存储 Bucket（存储桶名称）
	TencentCosBucket string `json:"tencent_cos_bucket"`

	// 腾讯云 COS 对象存储所属地域，例如：ap-guangzhou
	TencentCosRegion string `json:"tencent_cos_region"`

	// 腾讯云 COS 对象存储预签名有效期，单位：秒
	TencentCosExpired time.Duration `json:"tencent_cos_expired"`

	// 腾讯云 COS 对象存储上传大小限制，单位：KB
	TencentCosLimit int `json:"tencent_cos_limit"`

	// 办公平台
	// 飞书
	OfficePlatform string `json:"office_platform"`

	// 飞书应用 ID
	FeishuAppId string `json:"feishu_app_id"`

	// 飞书应用密钥
	FeishuAppSecret string `json:"feishu_app_secret"`

	// 飞书事件订阅安全校验数据密钥
	FeishuEncryptKey string `json:"feishu_encrypt_key"`

	// 飞书事件订阅验证令牌
	FeishuVerificationToken string `json:"feishu_verification_token"`

	// 第三方免登授权码跳转地址
	RedirectUrl string `json:"redirect_url"`

	// 公共电子邮件服务 SMTP 地址
	EmailHost string `json:"email_host"`

	// SMTP 端口号（SSL）
	EmailPort string `json:"email_port"`

	// 公共邮箱用户，例如：support@example.com
	EmailUsername string `json:"email_username"`

	// 公共邮箱用户密码
	EmailPassword string `json:"email_password"`

	// 开放服务地址
	OpenapiUrl string `json:"openapi_url"`

	// 开放服务应用认证 Key
	// API 网关应用认证方式 https://cloud.tencent.com/document/product/628/55088
	OpenapiKey string `json:"openapi_key"`

	// 开放服务应用认证密钥
	OpenapiSecret string `json:"openapi_secret"`
}

// SetValues 设置动态配置
func SetValues(object nats.ObjectStore) (values *Values, err error) {
	var b []byte
	if b, err = object.GetBytes("values"); err != nil {
		if err == nats.ErrObjectNotFound {
			v := Values{
				UserSessionExpire:    time.Hour,
				UserLoginFailedTimes: 5,
				UserLockTime:         time.Minute * 15,
				IpLoginFailedTimes:   10,
				IpWhitelist:          []string{},
				IpBlacklist:          []string{},
				PasswordStrength:     1,
				PasswordExpire:       365,
				TencentCosExpired:    time.Second * 300,
				TencentCosLimit:      5120,
				EmailPort:            "465",
			}
			if b, err = jsoniter.Marshal(v); err != nil {
				return
			}
			if _, err = object.PutBytes("values", b); err != nil {
				return
			}
			return &v, nil
		} else {
			return
		}
	}
	if err = jsoniter.Unmarshal(b, &values); err != nil {
		return
	}
	return
}

// WatchValues 监听配置
func WatchValues(object nats.ObjectStore, values *Values) (err error) {
	var watch nats.ObjectWatcher
	if watch, err = object.Watch(); err != nil {
		return
	}
	current := time.Now()
	for o := range watch.Updates() {
		if o == nil || o.ModTime.Unix() < current.Unix() {
			continue
		}
		if o.Name == "values" {
			var b []byte
			b, err = object.GetBytes("values")
			if err != nil {
				return
			}
			if err = jsoniter.Unmarshal(b, values); err != nil {
				// TODO: 配置同步异常提示
				return
			}
		}
	}
	return
}
