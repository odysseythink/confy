package confy

import (
	"fmt"
	"strings"
)

type iniCodec struct{}

func (iniCodec) Encode(v map[string]any) ([]byte, error) {
	var b strings.Builder
	for key, val := range v {
		b.WriteString(fmt.Sprintf("%s=%v\n", key, val))
	}
	return []byte(b.String()), nil
}

func (iniCodec) Decode(b []byte, v map[string]any) error {
	var section string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line[1 : len(line)-1])
			continue
		}

		sep := "="
		if strings.Contains(line, ":") && !strings.Contains(line, "=") {
			sep = ":"
		}
		key, value, found := strings.Cut(line, sep)
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if section != "" {
			key = section + "." + key
		}
		v[strings.ToLower(key)] = value
	}
	return nil
}
