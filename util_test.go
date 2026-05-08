package confy

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigParseError(t *testing.T) {
	err := ConfigParseError{err: os.ErrNotExist}
	got := err.Error()
	want := "While parsing config: file does not exist"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
	if err.Unwrap() != os.ErrNotExist {
		t.Fatalf("expected os.ErrNotExist, got %v", err.Unwrap())
	}
}

func TestConfigMarshalError(t *testing.T) {
	err := ConfigMarshalError{err: os.ErrInvalid}
	got := err.Error()
	want := "While marshaling config: invalid argument"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestUnsupportedConfigError(t *testing.T) {
	err := UnsupportedConfigError("xml")
	got := err.Error()
	want := `Unsupported Config Type "xml"`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestConfigFileNotFoundError(t *testing.T) {
	err := ConfigFileNotFoundError{name: "app", locations: "/etc"}
	got := err.Error()
	want := `Config File "app" Not Found in "/etc"`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestConfigFileAlreadyExistsError(t *testing.T) {
	err := ConfigFileAlreadyExistsError("/etc/app.yaml")
	got := err.Error()
	want := `Config File "/etc/app.yaml" Already Exists`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestUnsupportedRemoteProviderError(t *testing.T) {
	err := UnsupportedRemoteProviderError("redis")
	got := err.Error()
	want := `Unsupported Remote Provider Type "redis"`
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRemoteConfigError(t *testing.T) {
	err := RemoteConfigError("connection refused")
	got := err.Error()
	want := "Remote Configurations Error: connection refused"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestToCaseInsensitiveValue(t *testing.T) {
	m := map[string]any{"Foo": "bar", "Baz": map[string]any{"Qux": 1}}
	result := toCaseInsensitiveValue(m)
	rm, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result)
	}
	if rm["foo"] != "bar" {
		t.Fatalf("expected bar, got %v", rm["foo"])
	}
	if _, ok := rm["Foo"]; ok {
		t.Fatal("expected original key to be removed")
	}

	m2 := map[any]any{"Foo": "bar"}
	result2 := toCaseInsensitiveValue(m2)
	rm2, ok := result2.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result2)
	}
	if rm2["foo"] != "bar" {
		t.Fatalf("expected bar, got %v", rm2["foo"])
	}

	// non-map value should pass through
	if toCaseInsensitiveValue("hello") != "hello" {
		t.Fatal("expected string to pass through")
	}
}

func TestCopyAndInsensitiviseMap(t *testing.T) {
	m := map[string]any{
		"Top": map[string]any{
			"Nested": map[string]any{
				"Deep": "value",
			},
		},
	}
	nm := copyAndInsensitiviseMap(m)
	if nm["top"].(map[string]any)["nested"].(map[string]any)["deep"] != "value" {
		t.Fatalf("unexpected value: %v", nm)
	}
	if _, ok := nm["Top"]; ok {
		t.Fatal("expected original key to be removed")
	}

	// map[any]any nested
	m2 := map[string]any{
		"A": map[any]any{"B": 1},
	}
	nm2 := copyAndInsensitiviseMap(m2)
	if nm2["a"].(map[string]any)["b"] != 1 {
		t.Fatalf("unexpected value: %v", nm2)
	}
}

func TestInsensitiviseVal(t *testing.T) {
	// map[any]any
	m := map[any]any{"Key": "value"}
	v := insensitiviseVal(m)
	sm, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", v)
	}
	if sm["key"] != "value" {
		t.Fatalf("expected value, got %v", sm["key"])
	}

	// map[string]any
	m2 := map[string]any{"Key": "value2"}
	v2 := insensitiviseVal(m2)
	sm2, ok := v2.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", v2)
	}
	if sm2["key"] != "value2" {
		t.Fatalf("expected value2, got %v", sm2["key"])
	}

	// []any
	a := []any{map[string]any{"Key": "value3"}}
	v3 := insensitiviseVal(a)
	sa, ok := v3.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", v3)
	}
	if sa[0].(map[string]any)["key"] != "value3" {
		t.Fatalf("expected value3, got %v", sa[0])
	}

	// primitive
	if insensitiviseVal(42) != 42 {
		t.Fatal("expected primitive to pass through")
	}
}

func TestAbsPathify(t *testing.T) {
	logger := slog.New(&discardHandler{})

	// relative path
	p := absPathify(logger, "./foo")
	if p == "" {
		t.Fatal("expected non-empty path")
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("expected absolute path, got %s", p)
	}

	// $HOME
	home := userHomeDir()
	if home != "" {
		p2 := absPathify(logger, "$HOME")
		if p2 != home {
			t.Fatalf("expected %s, got %s", home, p2)
		}
		p3 := absPathify(logger, "$HOME/subdir")
		if p3 != filepath.Join(home, "subdir") {
			t.Fatalf("expected %s, got %s", filepath.Join(home, "subdir"), p3)
		}
	}

	// ExpandEnv
	t.Setenv("TEST_CONFY_PATH", "/tmp/testconfy")
	p4 := absPathify(logger, "$TEST_CONFY_PATH")
	if p4 != "/tmp/testconfy" {
		t.Fatalf("expected /tmp/testconfy, got %s", p4)
	}
}

func TestUserHomeDir(t *testing.T) {
	home := userHomeDir()
	if home == "" {
		t.Fatal("expected non-empty home dir")
	}
}

func TestSafeMul(t *testing.T) {
	if safeMul(2, 3) != 6 {
		t.Fatal("expected 6")
	}
	if safeMul(0, 100) != 0 {
		t.Fatal("expected 0")
	}
	// overflow case
	big := uint(1 << 63)
	if safeMul(big, 4) != 0 {
		t.Fatal("expected 0 on overflow")
	}
}

func TestParseSizeInBytes(t *testing.T) {
	tests := []struct {
		input string
		want  uint
	}{
		{"1GB", 1 << 30},
		{"12 mb", 12 << 20},
		{"1024", 1024},
		{"1KB", 1 << 10},
		{"1kb", 1 << 10},
		{"1Gb", 1 << 30},
		{"1b", 0},
		{"", 0},
		{"0", 0},
		{"5 MB", 5 << 20},
		{"123B", 123},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSizeInBytes(tt.input)
			if got != tt.want {
				t.Fatalf("parseSizeInBytes(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeepSearch(t *testing.T) {
	m := map[string]any{}
	result := deepSearch(m, []string{"a", "b", "c"})
	result["d"] = "value"

	if m["a"].(map[string]any)["b"].(map[string]any)["c"].(map[string]any)["d"] != "value" {
		t.Fatalf("unexpected structure: %v", m)
	}

	// replacing existing value with map
	m2 := map[string]any{"x": "old"}
	result2 := deepSearch(m2, []string{"x", "y"})
	result2["z"] = "new"
	if m2["x"].(map[string]any)["y"].(map[string]any)["z"] != "new" {
		t.Fatalf("unexpected structure: %v", m2)
	}
}
