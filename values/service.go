package values

import (
	"errors"
	"github.com/nats-io/nats.go"
	"github.com/thoas/go-funk"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/weplanx/go-wpx/cipher"
)

type Service struct {
	KeyValue nats.KeyValue
	Cipher   *cipher.Cipher
	Values   *Values
}

func (x *Service) Fetch() (data map[string]interface{}, err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		if !errors.Is(err, nats.ErrKeyNotFound) {
			return
		}
		if err = x.Update(x.Values); err != nil {
			return
		}
		return
	}
	var b []byte
	if b, err = x.Cipher.Decode(string(entry.Value())); err != nil {
		return
	}
	if err = msgpack.Unmarshal(b, &x.Values); err != nil {
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
	if err = x.Fetch(); err != nil {
		return
	}
	for path, value := range data {
		if err = funk.Set(&x.Values, value, path); err != nil {
			return
		}
	}
	return x.Update(values)
}

func (x *Service) Get(keys []string) (values map[string]interface{}, err error) {
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	var b []byte
	if b, err = x.Cipher.Decode(string(entry.Value())); err != nil {
		return
	}
	if err = msgpack.Unmarshal(b, &values); err != nil {
		return
	}
	for k, v := range values {
		if len(keys) != 0 && !funk.Contains(keys, k) {
			delete(values, k)
			continue
		}
		if funk.Contains(SECRET, k) {
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
	var entry nats.KeyValueEntry
	if entry, err = x.KeyValue.Get("values"); err != nil {
		return
	}
	var b []byte
	if b, err = x.Cipher.Decode(string(entry.Value())); err != nil {
		return
	}
	var values map[string]interface{}
	if err = msgpack.Unmarshal(b, &values); err != nil {
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
