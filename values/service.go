package values

import (
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/go-wpx/cipher"
	"reflect"
)

type Service struct {
	KeyValue nats.KeyValue
	Cipher   *cipher.Cipher
	Values   *Values
}

func (x *Service) Fetch(values map[string]interface{}) (err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		if errors.Is(err, nats.ErrKeyNotFound) {
			v := reflect.ValueOf(*x.Values)
			typ := v.Type()
			for i := 0; i < v.NumField(); i++ {
				values[typ.Field(i).Name] = v.Field(i).Interface()
			}
			return x.Update(values)
		}
		return
	}
	var b []byte
	if b, err = x.Cipher.Decode(string(entry.Value())); err != nil {
		return
	}
	if err = msgpack.Unmarshal(b, &values); err != nil {
		return
	}
	return
}

//type SyncOption struct {
//	Updated chan *Values
//	Err     chan error
//}
//
//func (x *Service) Sync(option *SyncOption) (err error) {
//	if err = x.Load(); err != nil {
//		return
//	}
//	current := time.Now()
//	var watch nats.ObjectWatcher
//	watch, err = x.Object.Watch()
//	for entry := range watch.Updates() {
//		if entry.Name != "values" {
//			continue
//		}
//		if entry == nil || entry.Created().Unix() < current.Unix() {
//			continue
//		}
//		if err = sonic.Unmarshal(entry.Value(), x.Values); err != nil {
//			if option != nil && option.Err != nil {
//				option.Err <- err
//			}
//			return
//		}
//		if option != nil && option.Updated != nil {
//			option.Updated <- x.Values
//		}
//	}
//
//	return
//}

func (x *Service) Set(data map[string]interface{}) (err error) {
	var values map[string]interface{}
	if err = x.Fetch(values); err != nil {
		return
	}
	for key, value := range data {
		values[key] = value
	}
	return x.Update(values)
}

func (x *Service) Get(keys []string) (values map[string]interface{}, err error) {
	if err = x.Fetch(values); err != nil {
		return
	}
	contains := make(map[string]bool)
	for _, v := range keys {
		contains[v] = true
	}
	for k, v := range values {
		if len(keys) != 0 && !contains[k] {
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

func (x *Service) Remove(key string) (err error) {
	var values map[string]interface{}
	if err = x.Fetch(values); err != nil {
		return
	}
	delete(values, key)
	return x.Update(values)
}

func (x *Service) Update(values interface{}) (err error) {
	var b []byte
	if b, err = msgpack.Marshal(values); err != nil {
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
