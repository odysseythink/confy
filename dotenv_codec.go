package confy

import (
	"fmt"
	"strings"
)

type dotenvCodec struct{}

func (dotenvCodec) Encode(v map[string]any) ([]byte, error) {
	var b strings.Builder
	for key, val := range v {
		b.WriteString(fmt.Sprintf("%s=%v\n", key, val))
	}
	return []byte(b.String()), nil
}

func (dotenvCodec) Decode(b []byte, v map[string]any) error {
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		v[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return nil
}
