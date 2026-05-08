package confy

import (
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	v := New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/nonexistent")

	_, err := v.findConfigFile()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestSearchInPath(t *testing.T) {
	v := New()
	v.SetConfigName("app")
	v.SetConfigType("json")

	m := &mockFS{files: map[string]string{
		"/etc/app.json": `{"key":"value"}`,
	}}
	v.SetFS(m)
	v.AddConfigPath("/etc")

	file, err := v.findConfigFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file != "/etc/app.json" {
		t.Fatalf("expected /etc/app.json, got %s", file)
	}
}
