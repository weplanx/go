package values

import (
	"github.com/bytedance/sonic"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/go/cipher"
	"github.com/weplanx/go/help"
	"reflect"
	"time"
)

type Service struct {
	Type     reflect.Type
	KeyValue nats.KeyValue
	Cipher   *cipher.Cipher
}

func (x *Service) Fetch(v interface{}) (err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	var b []byte
	if b, err = x.Cipher.Decode(string(entry.Value())); err != nil {
		return
	}
	if err = sonic.Unmarshal(b, v); err != nil {
		return
	}
	return
}

func (x *Service) Sync(v interface{}, update chan interface{}) (err error) {
	if err = x.Fetch(v); err != nil {
		return
	}
	if update != nil {
		update <- v
	}
	current := time.Now()
	var watch nats.KeyWatcher
	watch, err = x.KeyValue.Watch("values")
	for entry := range watch.Updates() {
		if entry == nil || entry.Created().Unix() < current.Unix() {
			continue
		}
		if err = x.Fetch(v); err != nil {
			return
		}
		if update != nil {
			update <- v
		}
	}
	return
}

func (x *Service) Set(update map[string]interface{}) (err error) {
	var values map[string]interface{}
	if err = x.Fetch(&values); err != nil {
		return
	}
	for key, value := range update {
		values[key] = value
	}
	return x.Update(values)
}

func (x *Service) Get(keys ...string) (data map[string]interface{}, err error) {
	if err = x.Fetch(&data); err != nil {
		return
	}
	contains := make(map[string]bool)
	for _, v := range keys {
		contains[v] = true
	}
	for key, value := range data {
		if len(keys) != 0 && !contains[key] || help.IsEmpty(value) {
			delete(data, key)
			continue
		}
		secret := false
		if field, ok := x.Type.FieldByName(key); ok {
			secret = field.Tag.Get("secret") == "*"
		}
		if secret {
			data[key] = "*"
		}
	}
	return
}

func (x *Service) Remove(keys ...string) (err error) {
	var values map[string]interface{}
	if err = x.Fetch(&values); err != nil {
		return
	}
	for _, key := range keys {
		delete(values, key)
	}
	return x.Update(values)
}

func (x *Service) Update(data interface{}) (err error) {
	var b []byte
	if b, err = sonic.Marshal(data); err != nil {
		return
	}
	var ciphertext string
	if ciphertext, err = x.Cipher.Encode(b); err != nil {
		return
	}
	if _, err = x.KeyValue.PutString("values", ciphertext); err != nil {
		return
	}
	return
}
