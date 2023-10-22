package values

import (
	"github.com/nats-io/nats.go"
	"github.com/weplanx/go/cipher"
	"reflect"
	"time"
)

func New(options ...Option) *Service {
	x := new(Service)
	for _, v := range options {
		v(x)
	}
	return x
}

type Option func(x *Service)

func SetKeyValue(v nats.KeyValue) Option {
	return func(x *Service) {
		x.KeyValue = v
	}
}

func SetCipher(v *cipher.Cipher) Option {
	return func(x *Service) {
		x.Cipher = v
	}
}

func SetType(v reflect.Type) Option {
	return func(x *Service) {
		x.Type = v
	}
}

type DynamicValues struct {
	// session period (seconds)
	// User inactivity for 1 hour, session will end
	SessionTTL time.Duration `yaml:"session_ttl"`
	// login lockout time
	// Locked for 15 minutes
	LoginTTL time.Duration `yaml:"login_ttl"`
	// Maximum number of login failures for a user
	// If you fail to log in 5 times consecutively within a limited time (lockout time),
	// your account will be locked
	LoginFailures int64 `yaml:"login_failures"`
	// Maximum number of login failures for the user's host IP
	// If the same host IP fails to log in 10 times continuously, the IP will be locked (period is the login_ttl)
	IpLoginFailures int64 `yaml:"ip_login_failures"`
	// IP whitelist
	// Whitelisting IPs does not restrict login failure lockouts
	IpWhitelist []string `yaml:"ip_whitelist"`
	// IP blacklist
	// will ban all access
	IpBlacklist []string `yaml:"ip_blacklist"`
	// password strength
	// 0: unlimited
	// 1: uppercase and lowercase letters
	// 2: uppercase and lowercase letters, numbers
	// 3: uppercase and lowercase letters, numbers, special characters
	PwdStrategy int `yaml:"pwd_strategy"`
	// password validity period
	// After the password expires, it is mandatory to change the password, 0: permanently valid
	PwdTTL time.Duration `yaml:"pwd_ttl"`
	// Public Cloud
	// Supported: Tencent Cloud `tencent`
	// Plan: AWS `aws`
	Cloud string `yaml:"cloud"`
	// Tencent Cloud API Secret Id
	// It is recommended to use CAM to assign the required permissions
	TencentSecretId string `yaml:"tencent_secret_id"`
	// Tencent Cloud API Secret Key
	TencentSecretKey string `yaml:"tencent_secret_key" secret:"*"`
	// Tencent Cloud COS bucket name
	TencentCosBucket string `yaml:"tencent_cos_bucket"`
	// Tencent Cloud COS bucket region, for example: ap-guangzhou
	TencentCosRegion string `yaml:"tencent_cos_region"`
	// Tencent Cloud COS bucket pre-signature validity period, unit: second
	TencentCosExpired int64 `yaml:"tencent_cos_expired"`
	// Tencent Cloud COS bucket upload size limit, unit: KB
	TencentCosLimit int64 `yaml:"tencent_cos_limit"`
	// Office collaboration platform
	Collaboration string `yaml:"collaboration"`
	// Enterprise Collaboration
	// Lark App ID
	LarkAppId string `yaml:"lark_app_id"`
	// Lark application key
	LarkAppSecret string `yaml:"lark_app_secret" secret:"*"`
	// Lark event subscription security verification data key
	LarkEncryptKey string `yaml:"lark_encrypt_key" secret:"*"`
	// Lark Event Subscription Verification Token
	LarkVerificationToken string `yaml:"lark_verification_token" secret:"*"`
	// Third-party registration-free authorization code redirection address
	RedirectUrl string `yaml:"redirect_url"`
	// Public email service SMTP address
	EmailHost string `yaml:"email_host"`
	// Public email SMTP port number (SSL)
	EmailPort int `yaml:"email_port"`
	// Public email username
	EmailUsername string `yaml:"email_username"`
	// Public email password
	EmailPassword string `yaml:"email_password" secret:"*"`
	// ApiGateway url
	ApiGatewayUrl string `yaml:"api_gateway_url"`
	// Openapi application authentication key
	// API gateway application authentication https://cloud.tencent.com/document/product/628/55088
	ApiGatewayKey string `yaml:"api_gateway_key"`
	// Openapi Application Authentication Secret
	ApiGatewaySecret string `yaml:"api_gateway_secret" secret:"*"`
	// RestControls
	RestControls map[string]*RestControl `yaml:"rest_controls"`
	// Rest Txn Timeout
	RestTxnTimeout time.Duration `yaml:"rest_txn_timeout"`
}

type RestControl struct {
	Keys       []string `yaml:"keys"`
	Sensitives []string `yaml:"sensitives"`
	Status     bool     `yaml:"status"`
	Event      bool     `yaml:"event"`
}
