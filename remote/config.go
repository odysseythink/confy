package remote

import (
	"bytes"
	"io"

	goetcdv3 "go.etcd.io/etcd/client/v3"
)

type configManager struct {
	keystore []byte
	store    Store
}

// A ConfigManager retrieves and decrypts configuration from a key/value store.
type ConfigManager interface {
	Get(key string) ([]byte, error)
	List(key string) (KVPairs, error)
	Set(key string, value []byte) error
	Watch(key string, stop chan bool) <-chan *Response
}

type standardConfigManager struct {
	store Store
}

func NewStandardConfigManager(client Store) (ConfigManager, error) {
	return standardConfigManager{client}, nil
}

func NewConfigManager(client Store, keystore io.Reader) (ConfigManager, error) {
	bytes, err := io.ReadAll(keystore)
	if err != nil {
		return nil, err
	}
	return configManager{bytes, client}, nil
}

// NewStandardEtcdV3ConfigManager returns a new ConfigManager backed by etcdv3.
func NewStandardEtcdV3ConfigManager(machines []string) (ConfigManager, error) {
	return NewStandardEtcdV3ConfigManagerFromConfig(goetcdv3.Config{
		Endpoints: machines,
	})
}

// NewStandardEtcdV3ConfigManagerFromConfig returns a new ConfigManager backed by etcdv3.
func NewStandardEtcdV3ConfigManagerFromConfig(config goetcdv3.Config) (ConfigManager, error) {
	store, err := NewEtcdClientFromV3Config(config)
	if err != nil {
		return nil, err
	}

	return NewStandardConfigManager(store)
}

// NewEtcdV3ConfigManager returns a new ConfigManager backed by etcdv3.
// Data will be encrypted.
func NewEtcdV3ConfigManager(machines []string, keystore io.Reader) (ConfigManager, error) {
	store, err := NewEtcdClientV3(machines)
	if err != nil {
		return nil, err
	}
	return NewConfigManager(store, keystore)
}

// Get retrieves and decodes a secconf value stored at key.
func (c configManager) Get(key string) ([]byte, error) {
	value, err := c.store.Get(key)
	if err != nil {
		return nil, err
	}
	return secconfDecode(value, bytes.NewBuffer(c.keystore))
}

// Get retrieves a value stored at key.
// convenience function, no additional value provided over
// `etcdctl`
func (c standardConfigManager) Get(key string) ([]byte, error) {
	value, err := c.store.Get(key)
	if err != nil {
		return nil, err
	}
	return value, err
}

// List retrieves and decodes all secconf value stored under key.
func (c configManager) List(key string) (KVPairs, error) {
	list, err := c.store.List(key)
	retList := make(KVPairs, len(list))
	if err != nil {
		return nil, err
	}
	for i, kv := range list {
		retList[i] = &KVPair{}
		retList[i].Key = kv.Key
		retList[i].Value, err = secconfDecode(kv.Value, bytes.NewBuffer(c.keystore))
		if err != nil {
			return nil, err
		}
	}
	return retList, nil
}

// List retrieves all values under key.
// convenience function, no additional value provided over
// `etcdctl`
func (c standardConfigManager) List(key string) (KVPairs, error) {
	list, err := c.store.List(key)
	retList := make(KVPairs, len(list))
	if err != nil {
		return nil, err
	}
	for i, kv := range list {
		retList[i] = kv
	}
	return retList, err
}

// Set will put a key/value into the data store
// and encode it with secconf
func (c configManager) Set(key string, value []byte) error {
	encodedValue, err := secconfEncode(value, bytes.NewBuffer(c.keystore))
	if err == nil {
		err = c.store.Set(key, encodedValue)
	}
	return err
}

// Set will put a key/value into the data store
func (c standardConfigManager) Set(key string, value []byte) error {
	err := c.store.Set(key, value)
	return err
}

func (c configManager) Watch(key string, stop chan bool) <-chan *Response {
	resp := make(chan *Response, 0)
	backendResp := c.store.Watch(key, stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			case r := <-backendResp:
				if r.Error != nil {
					resp <- &Response{nil, r.Error}
					continue
				}
				value, err := secconfDecode(r.Value, bytes.NewBuffer(c.keystore))
				resp <- &Response{value, err}
			}
		}
	}()
	return resp
}

func (c standardConfigManager) Watch(key string, stop chan bool) <-chan *Response {
	resp := make(chan *Response, 0)
	backendResp := c.store.Watch(key, stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			case r := <-backendResp:
				if r.Error != nil {
					resp <- &Response{nil, r.Error}
					continue
				}
				resp <- &Response{r.Value, nil}
			}
		}
	}()
	return resp
}
