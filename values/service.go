package values

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
)

type Service struct {
	Object nats.ObjectStore
}

// Set 分发配置
func (x *Service) Set(data map[string]interface{}) (err error) {
	var b []byte
	if b, err = x.Object.GetBytes("values"); err != nil {
		return
	}
	var values map[string]interface{}
	if err = jsoniter.Unmarshal(b, &values); err != nil {
		return
	}
	for k, v := range data {
		values[k] = v
	}
	if b, err = jsoniter.Marshal(values); err != nil {
		return
	}
	if _, err = x.Object.PutBytes("values", b); err != nil {
		return
	}
	return
}

// Get 获取配置
func (x *Service) Get(keys []string) (data map[string]interface{}, err error) {
	var b []byte
	if b, err = x.Object.GetBytes("values"); err != nil {
		return
	}
	values := make(map[string]interface{})
	if err = jsoniter.Unmarshal(b, &values); err != nil {
		return
	}
	data = make(map[string]interface{})
	for _, key := range keys {
		if x.IsSecret(key) {
			value := values[key]
			if value == nil || value == "" {
				data[key] = "-"
			} else {
				data[key] = "*"
			}
		} else {
			data[key] = values[key]
		}
	}
	return
}

// IsSecret 是否为密文
func (x *Service) IsSecret(key string) bool {
	return funk.Contains([]string{
		"tencent_secret_key",
		"tencent_pulsar_token",
		"feishu_app_secret",
		"feishu_encrypt_key",
		"feishu_verification_token",
		"email_password",
		"openapi_secret",
	}, key)
}
