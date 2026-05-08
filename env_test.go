package confy

import (
	"os"
	"testing"
)

func TestSetEnvPrefix(t *testing.T) {
	Reset()
	SetEnvPrefix("myapp")
	if GetEnvPrefix() != "myapp" {
		t.Fatalf("expected myapp, got %s", GetEnvPrefix())
	}
}

func TestAutomaticEnv(t *testing.T) {
	Reset()
	SetEnvPrefix("confy")
	AutomaticEnv()
	os.Setenv("CONFY_TESTKEY", "from_env")
	defer os.Unsetenv("CONFY_TESTKEY")

	val := GetString("testkey")
	if val != "from_env" {
		t.Fatalf("expected from_env, got %s", val)
	}
}

func TestBindEnv(t *testing.T) {
	Reset()
	BindEnv("mykey", "CUSTOM_ENV_VAR")
	os.Setenv("CUSTOM_ENV_VAR", "bound_value")
	defer os.Unsetenv("CUSTOM_ENV_VAR")

	val := GetString("mykey")
	if val != "bound_value" {
		t.Fatalf("expected bound_value, got %s", val)
	}
}

func TestAllowEmptyEnv(t *testing.T) {
	Reset()
	AllowEmptyEnv(true)
	SetEnvPrefix("confy")
	AutomaticEnv()
	os.Setenv("CONFY_EMPTY", "")
	defer os.Unsetenv("CONFY_EMPTY")

	val := GetString("empty")
	if val != "" {
		t.Fatalf("expected empty string, got %s", val)
	}
}
