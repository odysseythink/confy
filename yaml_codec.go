package confy

import "gopkg.in/yaml.v3"

// Codec implements the encoding.Encoder and encoding.Decoder interfaces for YAML encoding.
type yamlCodec struct{}

func (yamlCodec) Encode(v map[string]any) ([]byte, error) {
	return yaml.Marshal(v)
}

func (yamlCodec) Decode(b []byte, v map[string]any) error {
	return yaml.Unmarshal(b, &v)
}
