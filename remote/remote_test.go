package remote

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/odysseythink/confy"
)

type fakeProvider struct {
	provider      string
	endpoint      string
	path          string
	secretKeyring string
}

func (f fakeProvider) Provider() string      { return f.provider }
func (f fakeProvider) Endpoint() string      { return f.endpoint }
func (f fakeProvider) Path() string          { return f.path }
func (f fakeProvider) SecretKeyring() string { return f.secretKeyring }

type mockStore struct {
	getValue []byte
	getErr   error
	listVal  KVPairs
	listErr  error
	setErr   error
}

func (m *mockStore) Get(key string) ([]byte, error) {
	return m.getValue, m.getErr
}

func (m *mockStore) List(key string) (KVPairs, error) {
	return m.listVal, m.listErr
}

func (m *mockStore) Set(key string, value []byte) error {
	return m.setErr
}

func (m *mockStore) Watch(key string, stop chan bool) <-chan *Response {
	ch := make(chan *Response, 1)
	go func() {
		select {
		case <-stop:
			return
		case ch <- &Response{Value: m.getValue, Error: m.getErr}:
		}
	}()
	return ch
}

func TestRemoteConfigProviderGetError(t *testing.T) {
	rc := remoteConfigProvider{}
	// Use invalid keyring for immediate error
	_, err := rc.Get(fakeProvider{provider: "etcd3", endpoint: "invalid:2379", path: "/config", secretKeyring: "/nonexistent"})
	if err == nil {
		t.Error("expected error for invalid keyring")
	}
}

func TestRemoteConfigProviderWatchError(t *testing.T) {
	rc := remoteConfigProvider{}
	_, err := rc.Watch(fakeProvider{provider: "etcd3", endpoint: "invalid:2379", path: "/config", secretKeyring: "/nonexistent"})
	if err == nil {
		t.Error("expected error for invalid keyring")
	}
}

func TestRemoteConfigProviderWatchChannelError(t *testing.T) {
	rc := remoteConfigProvider{}
	ch, quit := rc.WatchChannel(fakeProvider{provider: "etcd3", endpoint: "invalid:2379", path: "/config", secretKeyring: "/nonexistent"})
	if ch != nil || quit != nil {
		t.Error("expected nil channels for invalid keyring")
	}
}

func TestGetConfigManagerInvalidKeyring(t *testing.T) {
	_, err := getConfigManager(fakeProvider{provider: "etcd3", endpoint: "invalid:2379", secretKeyring: "/nonexistent"})
	if err == nil {
		t.Error("expected error for invalid keyring")
	}
}

func TestGetConfigManagerNoKeyring(t *testing.T) {
	// Without keyring, it tries to create etcd client (which may succeed even with invalid endpoint)
	// Just verify it doesn't panic
	_, _ = getConfigManager(fakeProvider{provider: "etcd3", endpoint: "invalid:2379"})
}

func TestRemoteConfigProviderGetSuccess(t *testing.T) {
	// Verify init set RemoteConfig
	if confy.RemoteConfig == nil {
		t.Error("RemoteConfig should be set by init")
	}
}

// Test standardConfigManager with mock store
func TestStandardConfigManagerGet(t *testing.T) {
	store := &mockStore{getValue: []byte("hello")}
	cm, _ := NewStandardConfigManager(store)
	val, err := cm.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(val, []byte("hello")) {
		t.Errorf("got %s", val)
	}
}

func TestStandardConfigManagerGetError(t *testing.T) {
	store := &mockStore{getErr: errors.New("fail")}
	cm, _ := NewStandardConfigManager(store)
	_, err := cm.Get("key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStandardConfigManagerList(t *testing.T) {
	store := &mockStore{listVal: KVPairs{{Key: "k", Value: []byte("v")}}}
	cm, _ := NewStandardConfigManager(store)
	pairs, err := cm.List("key")
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 1 || pairs[0].Key != "k" {
		t.Errorf("got %v", pairs)
	}
}

func TestStandardConfigManagerListError(t *testing.T) {
	store := &mockStore{listErr: errors.New("fail")}
	cm, _ := NewStandardConfigManager(store)
	_, err := cm.List("key")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStandardConfigManagerSet(t *testing.T) {
	store := &mockStore{}
	cm, _ := NewStandardConfigManager(store)
	if err := cm.Set("key", []byte("val")); err != nil {
		t.Fatal(err)
	}
}

func TestStandardConfigManagerSetError(t *testing.T) {
	store := &mockStore{setErr: errors.New("fail")}
	cm, _ := NewStandardConfigManager(store)
	if err := cm.Set("key", []byte("val")); err == nil {
		t.Error("expected error")
	}
}

func TestStandardConfigManagerWatch(t *testing.T) {
	store := &mockStore{getValue: []byte("hello")}
	cm, _ := NewStandardConfigManager(store)
	stop := make(chan bool)
	ch := cm.Watch("key", stop)
	resp := <-ch
	if !bytes.Equal(resp.Value, []byte("hello")) {
		t.Errorf("got %v", resp)
	}
	close(stop)
}

// Test configManager with mock store (requires valid keystore for secconf)
func TestConfigManagerGetInvalidKeystore(t *testing.T) {
	store := &mockStore{getValue: []byte("hello")}
	cm, err := NewConfigManager(store, strings.NewReader("invalid"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cm.Get("key")
	if err == nil {
		t.Error("expected error for invalid keystore decode")
	}
}

func TestNewConfigManagerReadError(t *testing.T) {
	_, err := NewConfigManager(&mockStore{}, &errorReader{})
	if err == nil {
		t.Error("expected error")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}


func TestConfigManagerListInvalidKeystore(t *testing.T) {
	store := &mockStore{listVal: KVPairs{{Key: "k", Value: []byte("hello")}}}
	cm, err := NewConfigManager(store, strings.NewReader("invalid"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = cm.List("key")
	if err == nil {
		t.Error("expected error for invalid keystore decode")
	}
}

func TestConfigManagerSetInvalidKeystore(t *testing.T) {
	store := &mockStore{}
	cm, err := NewConfigManager(store, strings.NewReader("invalid"))
	if err != nil {
		t.Fatal(err)
	}
	if err := cm.Set("key", []byte("val")); err == nil {
		t.Error("expected error for invalid keystore encode")
	}
}

func TestConfigManagerWatchInvalidKeystore(t *testing.T) {
	store := &mockStore{getValue: []byte("hello")}
	cm, err := NewConfigManager(store, strings.NewReader("invalid"))
	if err != nil {
		t.Fatal(err)
	}
	stop := make(chan bool)
	ch := cm.Watch("key", stop)
	resp := <-ch
	if resp.Error == nil {
		t.Error("expected error for invalid keystore decode")
	}
	close(stop)
}

func TestSecconfDecodeInvalid(t *testing.T) {
	_, err := secconfDecode([]byte("invalid-data"), strings.NewReader("invalid-key"))
	if err == nil {
		t.Error("expected error")
	}
}

func TestSecconfEncodeInvalid(t *testing.T) {
	_, err := secconfEncode([]byte("hello"), strings.NewReader("invalid-key"))
	if err == nil {
		t.Error("expected error")
	}
}
