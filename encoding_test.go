package confy

import (
	"testing"
)

func TestDotenvCodec(t *testing.T) {
	codec := dotenvCodec{}
	input := map[string]any{
		"KEY1": "value1",
		"KEY2": "value2",
	}
	b, err := codec.Encode(input)
	if err != nil {
		t.Fatal(err)
	}

	output := make(map[string]any)
	err = codec.Decode(b, output)
	if err != nil {
		t.Fatal(err)
	}

	if output["KEY1"] != "value1" {
		t.Fatalf("expected value1, got %v", output["KEY1"])
	}
}

func TestDotenvCodecDecode(t *testing.T) {
	codec := dotenvCodec{}
	data := []byte(`
# comment
KEY1=value1
KEY2 = value2
`)
	output := make(map[string]any)
	err := codec.Decode(data, output)
	if err != nil {
		t.Fatal(err)
	}
	if output["KEY1"] != "value1" || output["KEY2"] != "value2" {
		t.Fatalf("unexpected output: %v", output)
	}
}
