package confy

import (
	"github.com/pelletier/go-toml/v2"
)

type tomlCodec struct{}

func (tomlCodec) Encode(v map[string]any) ([]byte, error) {
	return toml.Marshal(v)
}

func (tomlCodec) Decode(b []byte, v map[string]any) error {
	return toml.Unmarshal(b, &v)
}
