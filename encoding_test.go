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

func TestIniCodecDecode(t *testing.T) {
	codec := iniCodec{}
	data := []byte(`
[database]
host = localhost
port = 3306

[server]
name = myapp
`)
	output := make(map[string]any)
	err := codec.Decode(data, output)
	if err != nil {
		t.Fatal(err)
	}
	if output["database.host"] != "localhost" {
		t.Fatalf("expected localhost, got %v", output["database.host"])
	}
	if output["database.port"] != "3306" {
		t.Fatalf("expected 3306, got %v", output["database.port"])
	}
	if output["server.name"] != "myapp" {
		t.Fatalf("expected myapp, got %v", output["server.name"])
	}
}

func TestTomlCodec(t *testing.T) {
	codec := tomlCodec{}
	input := map[string]any{
		"title": "Test",
		"owner": map[string]any{
			"name": "Tom",
		},
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

	if output["title"] != "Test" {
		t.Fatalf("expected Test, got %v", output["title"])
	}
}


func TestWithEncoderRegistry(t *testing.T) {
	reg := NewCodecRegistry()
	_ = reg.RegisterCodec("custom", jsonCodec{})
	v := NewWithOptions(WithEncoderRegistry(reg))
	enc, err := v.encoderRegistry.Encoder("custom")
	if err != nil {
		t.Fatal(err)
	}
	if enc == nil {
		t.Fatal("expected encoder")
	}
}

func TestWithDecoderRegistry(t *testing.T) {
	reg := NewCodecRegistry()
	_ = reg.RegisterCodec("custom", jsonCodec{})
	v := NewWithOptions(WithDecoderRegistry(reg))
	dec, err := v.decoderRegistry.Decoder("custom")
	if err != nil {
		t.Fatal(err)
	}
	if dec == nil {
		t.Fatal("expected decoder")
	}
}

func TestWithCodecRegistry(t *testing.T) {
	reg := NewCodecRegistry()
	_ = reg.RegisterCodec("custom", jsonCodec{})
	v := NewWithOptions(WithCodecRegistry(reg))
	enc, err := v.encoderRegistry.Encoder("custom")
	if err != nil {
		t.Fatal(err)
	}
	dec, err := v.decoderRegistry.Decoder("custom")
	if err != nil {
		t.Fatal(err)
	}
	if enc == nil || dec == nil {
		t.Fatal("expected codec")
	}
}

func TestWithCodecRegistryNil(t *testing.T) {
	v := NewWithOptions(WithCodecRegistry(nil))
	if v.encoderRegistry == nil {
		t.Fatal("expected default encoder registry")
	}
}

func TestDefaultCodecRegistryRegisterCodec(t *testing.T) {
	reg := NewCodecRegistry()
	err := reg.RegisterCodec("custom", jsonCodec{})
	if err != nil {
		t.Fatal(err)
	}
	codec, ok := reg.codec("custom")
	if !ok || codec == nil {
		t.Fatal("expected registered codec")
	}
}

func TestDefaultCodecRegistryEncoderNotFound(t *testing.T) {
	reg := NewCodecRegistry()
	_, err := reg.Encoder("unknown")
	if err == nil {
		t.Fatal("expected error for unknown encoder")
	}
}

func TestDefaultCodecRegistryDecoderNotFound(t *testing.T) {
	reg := NewCodecRegistry()
	_, err := reg.Decoder("unknown")
	if err == nil {
		t.Fatal("expected error for unknown decoder")
	}
}
