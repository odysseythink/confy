package confy

import (
	"encoding/json"
)

// jsonCodec implements the encoding.Encoder and encoding.Decoder interfaces for JSON encoding.
type jsonCodec struct{}

func (jsonCodec) Encode(v map[string]any) ([]byte, error) {
	// TODO: expose prefix and indent in the Codec as setting?
	return json.MarshalIndent(v, "", "  ")
}

func (jsonCodec) Decode(b []byte, v map[string]any) error {
	return json.Unmarshal(b, &v)
}
