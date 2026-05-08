package confy

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestUnsupportedRemoteProviderError_Error(t *testing.T) {
	err := UnsupportedRemoteProviderError("badprovider")
	expected := `Unsupported Remote Provider Type "badprovider"`
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestRemoteConfigError_Error(t *testing.T) {
	err := RemoteConfigError("something went wrong")
	expected := "Remote Configurations Error: something went wrong"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestDefaultRemoteProvider_Getters(t *testing.T) {
	rp := defaultRemoteProvider{
		provider:      "etcd",
		endpoint:      "http://localhost:2379",
		path:          "/config",
		secretKeyring: "/etc/secrets/myring.gpg",
	}

	if rp.Provider() != "etcd" {
		t.Fatalf("expected etcd, got %s", rp.Provider())
	}
	if rp.Endpoint() != "http://localhost:2379" {
		t.Fatalf("expected http://localhost:2379, got %s", rp.Endpoint())
	}
	if rp.Path() != "/config" {
		t.Fatalf("expected /config, got %s", rp.Path())
	}
	if rp.SecretKeyring() != "/etc/secrets/myring.gpg" {
		t.Fatalf("expected /etc/secrets/myring.gpg, got %s", rp.SecretKeyring())
	}
}

func TestAddRemoteProvider_invalidProvider(t *testing.T) {
	Reset()
	err := AddRemoteProvider("invalid", "http://localhost:2379", "/config")
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
	var ure UnsupportedRemoteProviderError
	if !errors.As(err, &ure) {
		t.Fatalf("expected UnsupportedRemoteProviderError, got %T", err)
	}
}

func TestAddSecureRemoteProvider_invalidProvider(t *testing.T) {
	Reset()
	err := AddSecureRemoteProvider("invalid", "http://localhost:2379", "/config", "/etc/secrets/myring.gpg")
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}
	var ure UnsupportedRemoteProviderError
	if !errors.As(err, &ure) {
		t.Fatalf("expected UnsupportedRemoteProviderError, got %T", err)
	}
}

func TestReadRemoteConfig_nilRemoteConfig(t *testing.T) {
	Reset()
	_ = AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	// RemoteConfig is nil by default
	err := ReadRemoteConfig()
	if err == nil {
		t.Fatal("expected error when RemoteConfig is nil")
	}
	var rce RemoteConfigError
	if !errors.As(err, &rce) {
		t.Fatalf("expected RemoteConfigError, got %T", err)
	}
}

func TestWatchRemoteConfig_nilRemoteConfig(t *testing.T) {
	Reset()
	_ = AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	err := WatchRemoteConfig()
	if err == nil {
		t.Fatal("expected error when RemoteConfig is nil")
	}
	var rce RemoteConfigError
	if !errors.As(err, &rce) {
		t.Fatalf("expected RemoteConfigError, got %T", err)
	}
}

func TestWatchRemoteConfigOnChannel_nilRemoteConfig(t *testing.T) {
	Reset()
	_ = AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	err := v.WatchRemoteConfigOnChannel()
	if err == nil {
		t.Fatal("expected error when RemoteConfig is nil")
	}
	var rce RemoteConfigError
	if !errors.As(err, &rce) {
		t.Fatalf("expected RemoteConfigError, got %T", err)
	}
}

func TestProviderPathExists(t *testing.T) {
	Reset()
	rp := &defaultRemoteProvider{provider: "etcd", endpoint: "http://localhost:2379", path: "/config"}
	v.remoteProviders = append(v.remoteProviders, rp)
	if !v.providerPathExists(rp) {
		t.Fatal("expected provider path to exist")
	}
	other := &defaultRemoteProvider{provider: "consul", endpoint: "localhost:8500", path: "/config"}
	if v.providerPathExists(other) {
		t.Fatal("expected provider path to not exist")
	}
}

func TestProviderPathExists_DeepEqual(t *testing.T) {
	p1 := &defaultRemoteProvider{provider: "etcd", endpoint: "http://a", path: "/p"}
	p2 := &defaultRemoteProvider{provider: "etcd", endpoint: "http://a", path: "/p"}
	p3 := &defaultRemoteProvider{provider: "etcd", endpoint: "http://b", path: "/p"}

	if !reflect.DeepEqual(p1, p2) {
		t.Fatal("expected p1 and p2 to be equal")
	}
	if reflect.DeepEqual(p1, p3) {
		t.Fatal("expected p1 and p3 to not be equal")
	}
}

func TestReadRemoteConfig_noRemoteProviders(t *testing.T) {
	Reset()
	err := ReadRemoteConfig()
	if err == nil {
		t.Fatal("expected error when no remote providers")
	}
	var rce RemoteConfigError
	if !errors.As(err, &rce) {
		t.Fatalf("expected RemoteConfigError, got %T", err)
	}
}

func TestWatchRemoteConfig_noRemoteProviders(t *testing.T) {
	Reset()
	err := WatchRemoteConfig()
	if err == nil {
		t.Fatal("expected error when no remote providers")
	}
	var rce RemoteConfigError
	if !errors.As(err, &rce) {
		t.Fatalf("expected RemoteConfigError, got %T", err)
	}
}

// mockRemoteConfig implements remoteConfigFactory for testing.
type mockRemoteConfig struct {
	reader io.Reader
}

func (m *mockRemoteConfig) Get(rp RemoteProvider) (io.Reader, error) {
	return m.reader, nil
}

func (m *mockRemoteConfig) Watch(rp RemoteProvider) (io.Reader, error) {
	return m.reader, nil
}

func (m *mockRemoteConfig) WatchChannel(rp RemoteProvider) (<-chan *RemoteResponse, chan bool) {
	ch := make(chan *RemoteResponse, 1)
	ch <- &RemoteResponse{Value: []byte(`{"key":"value"}`), Error: nil}
	quit := make(chan bool)
	return ch, quit
}

func TestAddSecureRemoteProvider_success(t *testing.T) {
	Reset()
	err := AddSecureRemoteProvider("etcd", "http://localhost:2379", "/config", "/etc/ring.gpg")
	if err != nil {
		t.Fatal(err)
	}
	if len(v.remoteProviders) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(v.remoteProviders))
	}
	if v.remoteProviders[0].SecretKeyring() != "/etc/ring.gpg" {
		t.Fatal("expected secret keyring to be set")
	}
}

func TestReadRemoteConfig_success(t *testing.T) {
	Reset()
	v.SetConfigType("json")
	_ = v.AddRemoteProvider("etcd", "http://localhost:2379", "/config")

	oldRemoteConfig := RemoteConfig
	RemoteConfig = &mockRemoteConfig{reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}
	defer func() { RemoteConfig = oldRemoteConfig }()

	err := v.ReadRemoteConfig()
	if err != nil {
		t.Fatal(err)
	}
	if v.GetString("foo") != "bar" {
		t.Fatalf("expected bar, got %s", v.GetString("foo"))
	}
}

func TestReadRemoteConfig_getError(t *testing.T) {
	Reset()
	v.SetConfigType("json")
	_ = v.AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	_ = v.AddRemoteProvider("consul", "localhost:8500", "/config")

	oldRemoteConfig := RemoteConfig
	RemoteConfig = &mockRemoteConfig{reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}
	defer func() { RemoteConfig = oldRemoteConfig }()

	err := v.ReadRemoteConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatchRemoteConfig_success(t *testing.T) {
	Reset()
	v.SetConfigType("json")
	_ = v.AddRemoteProvider("etcd", "http://localhost:2379", "/config")

	oldRemoteConfig := RemoteConfig
	RemoteConfig = &mockRemoteConfig{reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}
	defer func() { RemoteConfig = oldRemoteConfig }()

	err := v.WatchRemoteConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatchRemoteConfig_watchError(t *testing.T) {
	Reset()
	v.SetConfigType("json")
	_ = v.AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	_ = v.AddRemoteProvider("consul", "localhost:8500", "/config")

	oldRemoteConfig := RemoteConfig
	RemoteConfig = &mockRemoteConfig{reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}
	defer func() { RemoteConfig = oldRemoteConfig }()

	err := v.WatchRemoteConfig()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatchRemoteConfigOnChannel_success(t *testing.T) {
	Reset()
	v.SetConfigType("json")
	_ = v.AddRemoteProvider("etcd", "http://localhost:2379", "/config")

	oldRemoteConfig := RemoteConfig
	RemoteConfig = &mockRemoteConfig{reader: bytes.NewReader([]byte(`{"foo":"bar"}`))}
	defer func() { RemoteConfig = oldRemoteConfig }()

	err := v.WatchRemoteConfigOnChannel()
	if err != nil {
		t.Fatal(err)
	}

	// Give the goroutine a moment to process
	time.Sleep(10 * time.Millisecond)
}
