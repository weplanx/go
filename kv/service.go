package kv

import (
	"errors"
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
	"time"
)

type Service struct {
	*KV
}

// Load 载入配置
func (x *Service) Load() (err error) {
	var b []byte
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		if !errors.Is(err, nats.ErrKeyNotFound) {
			return
		}
		b, _ = sonic.Marshal(x.DynamicValues)
		if _, err = x.KeyValue.Put("values", b); err != nil {
			return
		}
	}

	if b == nil {
		b = entry.Value()
	}

	if err = sonic.Unmarshal(b, &x.DynamicValues); err != nil {
		return
	}

	return
}

type SyncOption struct {
	Updated chan *DynamicValues
	Err     chan error
}

// Sync 同步节点动态配置
func (x *Service) Sync(option *SyncOption) (err error) {
	if err = x.Load(); err != nil {
		return
	}

	current := time.Now()
	watch, _ := x.KeyValue.Watch("values")

	for entry := range watch.Updates() {
		if entry == nil || entry.Created().Unix() < current.Unix() {
			continue
		}
		// 同步动态配置
		if err = sonic.Unmarshal(entry.Value(), x.DynamicValues); err != nil {
			if option != nil && option.Err != nil {
				option.Err <- err
			}
			return
		}
		if option != nil && option.Updated != nil {
			option.Updated <- x.DynamicValues
		}
	}

	return
}

// Set 设置动态配置
func (x *Service) Set(update map[string]interface{}) (err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	var values map[string]interface{}
	if err = sonic.Unmarshal(entry.Value(), &values); err != nil {
		return
	}
	for k, v := range update {
		values[k] = v
	}
	return x.Update(values)
}

var SECRET = map[string]bool{
	"tencent_secret_key":        true,
	"feishu_app_secret":         true,
	"feishu_encrypt_key":        true,
	"feishu_verification_token": true,
	"email_password":            true,
	"openapi_secret":            true,
}

// Get 获取动态配置
func (x *Service) Get(keys map[string]int64) (values map[string]interface{}, err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	if err = sonic.Unmarshal(entry.Value(), &values); err != nil {
		return
	}
	for k, v := range values {
		if len(keys) != 0 && keys[k] != 1 {
			delete(values, k)
			continue
		}
		if SECRET[k] {
			if funk.IsEmpty(v) {
				values[k] = "-"
			} else {
				values[k] = "*"
			}
		}
	}
	return
}

// Remove 移除动态配置
func (x *Service) Remove(key string) (err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	var values map[string]interface{}
	if err = sonic.Unmarshal(entry.Value(), &values); err != nil {
		return
	}
	delete(values, key)
	return x.Update(values)
}

// Update 更新配置
func (x *Service) Update(values map[string]interface{}) (err error) {
	b, _ := sonic.Marshal(values)
	if _, err = x.KeyValue.Put("values", b); err != nil {
		return
	}
	return
}
