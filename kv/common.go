package kv

import (
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"time"
)

var Provides = wire.NewSet(
	wire.Struct(new(Controller), "*"),
	wire.Struct(new(Service), "*"),
)

type KV struct {
	Namespace     string
	KeyValue      nats.KeyValue
	DynamicValues *DynamicValues
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
	SessionTTL time.Duration `json:"session_ttl"`
	// login lockout time
	// Locked for 15 minutes
	LoginTTL time.Duration `json:"login_ttl"`
	// Maximum number of login failures for a user
	// If you fail to log in 5 times consecutively within a limited time (lockout time),
	// your account will be locked
	LoginFailures int64 `json:"login_failures"`
	// Maximum number of login failures for the user's host IP
	// If the same host IP fails to log in 10 times continuously, the IP will be locked (period is the login_ttl)
	IpLoginFailures int64 `json:"ip_login_failures"`
	// IP whitelist
	// Whitelisting IPs does not restrict login failure lockouts
	IpWhitelist []string `json:"ip_whitelist"`
	// IP blacklist
	// will ban all access
	IpBlacklist []string `json:"ip_blacklist"`
	// password strength
	// 0: unlimited
	// 1: uppercase and lowercase letters
	// 2: uppercase and lowercase letters, numbers
	// 3: uppercase and lowercase letters, numbers, special characters
	PwdStrategy int `json:"pwd_strategy"`
	// password validity period
	// After the password expires, it is mandatory to change the password, 0: permanently valid
	PwdTTL time.Duration `json:"pwd_ttl"`
	// Public Cloud
	// Supported: Tencent Cloud `tencent`
	// Plan: AWS `aws`, Alibaba Cloud `aliyun`
	Cloud string `json:"cloud"`
	// Tencent Cloud API Secret Id
	// It is recommended to use CAM to assign the required permissions
	TencentSecretId string `json:"tencent_secret_id"`
	// Tencent Cloud API Secret Key
	TencentSecretKey string `json:"tencent_secret_key,omitempty"`
	// Tencent Cloud COS bucket name
	TencentCosBucket string `json:"tencent_cos_bucket,omitempty"`
	// Tencent Cloud COS bucket region, for example: ap-guangzhou
	TencentCosRegion string `json:"tencent_cos_region"`
	// Tencent Cloud COS bucket pre-signature validity period, unit: second
	TencentCosExpired int `json:"tencent_cos_expired"`
	// Tencent Cloud COS bucket upload size limit, unit: KB
	TencentCosLimit int64 `json:"tencent_cos_limit"`
	// Enterprise Collaboration
	// Feishu App ID
	FeishuAppId string `json:"feishu_app_id"`
	// Feishu application key
	FeishuAppSecret string `json:"feishu_app_secret,omitempty"`
	// Feishu event subscription security verification data key
	FeishuEncryptKey string `json:"feishu_encrypt_key,omitempty"`
	// Feishu Event Subscription Verification Token
	FeishuVerificationToken string `json:"feishu_verification_token,omitempty"`
	// Third-party registration-free authorization code redirection address
	RedirectUrl string `json:"redirect_url"`
	// Public email service SMTP address
	EmailHost string `json:"email_host"`
	// Public email SMTP port number (SSL)
	EmailPort string `json:"email_port"`
	// Public email username
	EmailUsername string `json:"email_username"`
	// Public email password
	EmailPassword string `json:"email_password,omitempty"`
	// Openapi url
	OpenapiUrl string `json:"openapi_url"`
	// Openapi application authentication key
	// API gateway application authentication https://cloud.tencent.com/document/product/628/55088
	OpenapiKey string `json:"openapi_key"`
	// Openapi Application Authentication Secret
	OpenapiSecret string `json:"openapi_secret,omitempty"`
	// DSL Control Variables
	DSL map[string]*DSLOption `json:"dsl,omitempty"`
}

type DSLOption struct {
	Event bool
	Keys  []string
}

func New(options ...Option) *KV {
	x := new(KV)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *KV)

func SetNamespace(v string) Option {
	return func(x *KV) {
		x.Namespace = v
	}
}

func SetKeyValue(v nats.KeyValue) Option {
	return func(x *KV) {
		x.KeyValue = v
	}
}

func SetDynamicValues(v *DynamicValues) Option {
	return func(x *KV) {
		x.DynamicValues = v
	}
}
