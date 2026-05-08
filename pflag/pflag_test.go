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
