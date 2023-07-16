package values

import (
	"github.com/nats-io/nats.go"
	"github.com/weplanx/go/cipher"
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

func SetCipher(v *cipher.Cipher) Option {
	return func(x *Service) {
		x.Cipher = v
	}
}

func SetDynamicValues(v *DynamicValues) Option {
	return func(x *Service) {
		x.Values = v
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
	SessionTTL time.Duration
	// login lockout time
	// Locked for 15 minutes
	LoginTTL time.Duration
	// Maximum number of login failures for a user
	// If you fail to log in 5 times consecutively within a limited time (lockout time),
	// your account will be locked
	LoginFailures int64
	// Maximum number of login failures for the user's host IP
	// If the same host IP fails to log in 10 times continuously, the IP will be locked (period is the login_ttl)
	IpLoginFailures int64
	// IP whitelist
	// Whitelisting IPs does not restrict login failure lockouts
	IpWhitelist []string
	// IP blacklist
	// will ban all access
	IpBlacklist []string
	// password strength
	// 0: unlimited
	// 1: uppercase and lowercase letters
	// 2: uppercase and lowercase letters, numbers
	// 3: uppercase and lowercase letters, numbers, special characters
	PwdStrategy int
	// password validity period
	// After the password expires, it is mandatory to change the password, 0: permanently valid
	PwdTTL time.Duration
	// Public Cloud
	// Supported: Tencent Cloud `tencent`
	// Plan: AWS `aws`
	Cloud string
	// Tencent Cloud API Secret Id
	// It is recommended to use CAM to assign the required permissions
	TencentSecretId string
	// Tencent Cloud API Secret Key
	TencentSecretKey string `secret:"*"`
	// Tencent Cloud COS bucket name
	TencentCosBucket string
	// Tencent Cloud COS bucket region, for example: ap-guangzhou
	TencentCosRegion string
	// Tencent Cloud COS bucket pre-signature validity period, unit: second
	TencentCosExpired int64
	// Tencent Cloud COS bucket upload size limit, unit: KB
	TencentCosLimit int64
	// Enterprise Collaboration
	// Lark App ID
	LarkAppId string
	// Lark application key
	LarkAppSecret string `secret:"*"`
	// Lark event subscription security verification data key
	LarkEncryptKey string `secret:"*"`
	// Lark Event Subscription Verification Token
	LarkVerificationToken string `secret:"*"`
	// Third-party registration-free authorization code redirection address
	RedirectUrl string
	// Public email service SMTP address
	EmailHost string
	// Public email SMTP port number (SSL)
	EmailPort int
	// Public email username
	EmailUsername string
	// Public email password
	EmailPassword string `secret:"*"`
	// Openapi url
	OpenapiUrl string
	// Openapi application authentication key
	// API gateway application authentication https://cloud.tencent.com/document/product/628/55088
	OpenapiKey string
	// Openapi Application Authentication Secret
	OpenapiSecret string `secret:"*"`
	// RestControls
	RestControls   map[string]*RestControl
	RestTxnTimeout time.Duration
}

type RestControl struct {
	Keys   []string
	Status bool
	Event  bool
}
