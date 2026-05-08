package pflag

import (
	"testing"

	"github.com/odysseythink/confy"
	"github.com/spf13/pflag"
)

func TestBindPFlag(t *testing.T) {
	v := confy.New()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Int("port", 8080, "port number")
	_ = fs.Set("port", "9090")

	err := BindPFlag(v, "port", fs.Lookup("port"))
	if err != nil {
		t.Fatal(err)
	}

	val := v.GetInt("port")
	if val != 9090 {
		t.Fatalf("expected 9090, got %d", val)
	}
}

func TestBindPFlags(t *testing.T) {
	v := confy.New()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String("host", "localhost", "host name")
	fs.Int("port", 8080, "port number")
	_ = fs.Set("host", "example.com")
	_ = fs.Set("port", "9090")

	err := BindPFlags(v, fs)
	if err != nil {
		t.Fatal(err)
	}

	if v.GetString("host") != "example.com" {
		t.Fatalf("expected example.com, got %s", v.GetString("host"))
	}
	if v.GetInt("port") != 9090 {
		t.Fatalf("expected 9090, got %d", v.GetInt("port"))
	}
}

func TestBindPFlag_NilFlag(t *testing.T) {
	v := confy.New()
	err := BindPFlag(v, "missing", nil)
	if err == nil {
		t.Fatal("expected error for nil flag")
	}
}

func TestBindPFlags_EmptyFlagSet(t *testing.T) {
	v := confy.New()
	fs := pflag.NewFlagSet("empty", pflag.ContinueOnError)

	err := BindPFlags(v, fs)
	if err != nil {
		t.Fatal(err)
	}
}
