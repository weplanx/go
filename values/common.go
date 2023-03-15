package values

import (
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"time"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
)

func New(options ...Option) *Service {
	x := new(Service)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Service)

func SetNamespace(v string) Option {
	return func(x *Service) {
		x.Namespace = v
	}
}

func SetKeyValue(v nats.KeyValue) Option {
	return func(x *Service) {
		x.KeyValue = v
	}
}

func SetDynamicValues(v *DynamicValues) Option {
	return func(x *Service) {
		x.DynamicValues = v
	}
}

var DEFAULT = DynamicValues{
	SessionTTL:      time.Hour,
	LoginTTL:        time.Minute * 15,
	LoginFailures:   5,
	IpLoginFailures: 10,
	IpWhitelist:     []string{},
	IpBlacklist:     []string{},
	PwdStrategy:     1,
	PwdTTL:          time.Hour * 24 * 365,
}

type DynamicValues struct {
	// session period (seconds)
	// User inactivity for 1 hour, session will end
	SessionTTL time.Duration `msgpack:"session_ttl"`
	// login lockout time
	// Locked for 15 minutes
	LoginTTL time.Duration `msgpack:"login_ttl"`
	// Maximum number of login failures for a user
	// If you fail to log in 5 times consecutively within a limited time (lockout time),
	// your account will be locked
	LoginFailures int64 `msgpack:"login_failures"`
	// Maximum number of login failures for the user's host IP
	// If the same host IP fails to log in 10 times continuously, the IP will be locked (period is the login_ttl)
	IpLoginFailures int64 `msgpack:"ip_login_failures"`
	// IP whitelist
	// Whitelisting IPs does not restrict login failure lockouts
	IpWhitelist []string `msgpack:"ip_whitelist"`
	// IP blacklist
	// will ban all access
	IpBlacklist []string `msgpack:"ip_blacklist"`
	// password strength
	// 0: unlimited
	// 1: uppercase and lowercase letters
	// 2: uppercase and lowercase letters, numbers
	// 3: uppercase and lowercase letters, numbers, special characters
	PwdStrategy int `msgpack:"pwd_strategy"`
	// password validity period
	// After the password expires, it is mandatory to change the password, 0: permanently valid
	PwdTTL time.Duration `msgpack:"pwd_ttl"`
	// Public Cloud
	// Supported: Tencent Cloud `tencent`
	// Plan: AWS `aws`, Alibaba Cloud `aliyun`
	Cloud string `msgpack:"cloud"`
	// Tencent Cloud API Secret Id
	// It is recommended to use CAM to assign the required permissions
	TencentSecretId string `msgpack:"tencent_secret_id"`
	// Tencent Cloud API Secret Key
	TencentSecretKey string `msgpack:"tencent_secret_key,omitempty"`
	// Tencent Cloud COS bucket name
	TencentCosBucket string `msgpack:"tencent_cos_bucket,omitempty"`
	// Tencent Cloud COS bucket region, for example: ap-guangzhou
	TencentCosRegion string `msgpack:"tencent_cos_region"`
	// Tencent Cloud COS bucket pre-signature validity period, unit: second
	TencentCosExpired int `msgpack:"tencent_cos_expired"`
	// Tencent Cloud COS bucket upload size limit, unit: KB
	TencentCosLimit int64 `msgpack:"tencent_cos_limit"`
	// Enterprise Collaboration
	// Feishu App ID
	FeishuAppId string `msgpack:"feishu_app_id"`
	// Feishu application key
	FeishuAppSecret string `msgpack:"feishu_app_secret,omitempty"`
	// Feishu event subscription security verification data key
	FeishuEncryptKey string `msgpack:"feishu_encrypt_key,omitempty"`
	// Feishu Event Subscription Verification Token
	FeishuVerificationToken string `msgpack:"feishu_verification_token,omitempty"`
	// Third-party registration-free authorization code redirection address
	RedirectUrl string `msgpack:"redirect_url"`
	// Public email service SMTP address
	EmailHost string `msgpack:"email_host"`
	// Public email SMTP port number (SSL)
	EmailPort string `msgpack:"email_port"`
	// Public email username
	EmailUsername string `msgpack:"email_username"`
	// Public email password
	EmailPassword string `msgpack:"email_password,omitempty"`
	// Openapi url
	OpenapiUrl string `msgpack:"openapi_url"`
	// Openapi application authentication key
	// API gateway application authentication https://cloud.tencent.com/document/product/628/55088
	OpenapiKey string `msgpack:"openapi_key"`
	// Openapi Application Authentication Secret
	OpenapiSecret string `msgpack:"openapi_secret,omitempty"`
	// Resources Control Variables
	Resources map[string]*ResourcesOption `msgpack:"resources,omitempty"`
}

type ResourcesOption struct {
	Event bool     `msgpack:"event"`
	Keys  []string `msgpack:"keys"`
}
