package cast

import (
	"reflect"
	"testing"
	"time"
)

func TestToStringMapE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]any
		wantErr  bool
	}{
		{"map[string]any", map[string]any{"a": 1}, map[string]any{"a": 1}, false},

		{"map[any]any", map[any]any{"a": 1}, map[string]any{"a": 1}, false},
		{"string JSON", `{"a":1}`, map[string]any{"a": float64(1)}, false},
		{"nil", nil, nil, true},
		{"unsupported", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringMapE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToStringMapStringE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]string
		wantErr  bool
	}{
		{"map[string]string", map[string]string{"a": "1"}, map[string]string{"a": "1"}, false},
		{"map[string]any", map[string]any{"a": 1}, map[string]string{"a": "1"}, false},
		{"map[any]string", map[any]string{"a": "1"}, map[string]string{"a": "1"}, false},
		{"map[any]any", map[any]any{"a": 1}, map[string]string{"a": "1"}, false},
		{"string JSON", `{"a":"1"}`, map[string]string{"a": "1"}, false},
		{"nil", nil, nil, true},
		{"unsupported", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringMapStringE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToStringMapStringSliceE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string][]string
		wantErr  bool
	}{
		{"map[string][]string", map[string][]string{"a": {"1"}}, map[string][]string{"a": {"1"}}, false},
		{"map[string][]any", map[string][]any{"a": {1}}, map[string][]string{"a": {"1"}}, false},
		{"map[string]string", map[string]string{"a": "1"}, map[string][]string{"a": {"1"}}, false},
		{"map[string]any with []any", map[string]any{"a": []any{1}}, map[string][]string{"a": {"1"}}, false},
		{"map[string]any with string", map[string]any{"a": "1"}, map[string][]string{"a": {"1"}}, false},
		{"map[string]any with []string", map[string]any{"a": []string{"1"}}, map[string][]string{"a": {"1"}}, false},
		{"map[any][]string", map[any][]string{"a": {"1"}}, map[string][]string{"a": {"1"}}, false},
		{"map[any]string", map[any]string{"a": "1"}, map[string][]string{"a": {"1"}}, false},
		{"map[any][]any", map[any][]any{"a": {1}}, map[string][]string{"a": {"1"}}, false},
		{"map[any]any", map[any]any{"a": []any{1}}, map[string][]string{"a": {"1"}}, false},
		{"string JSON", `{"a":["1"]}`, map[string][]string{"a": {"1"}}, false},
		{"nil", nil, nil, true},
		{"unsupported", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringMapStringSliceE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}

	// Test error cases for map[any]any with bad key/value
	_, err := ToStringMapStringSliceE(map[any]any{struct{}{}: []any{"1"}})
	if err == nil {
		t.Error("expected error for unstringable key")
	}
	_, err = ToStringMapStringSliceE(map[any]any{"a": struct{}{}})
	if err == nil {
		t.Error("expected error for unsliceable value")
	}
}

func TestToStringMapBoolE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]bool
		wantErr  bool
	}{
		{"map[string]bool", map[string]bool{"a": true}, map[string]bool{"a": true}, false},
		{"map[string]any", map[string]any{"a": true}, map[string]bool{"a": true}, false},
		{"map[any]bool", map[any]bool{"a": true}, map[string]bool{"a": true}, false},
		{"map[any]any", map[any]any{"a": true}, map[string]bool{"a": true}, false},
		{"string JSON", `{"a":true}`, map[string]bool{"a": true}, false},
		{"nil", nil, nil, true},
		{"unsupported", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringMapBoolE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToStringMapIntE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected map[string]int
		wantErr  bool
	}{
		{"map[string]int", map[string]int{"a": 1}, map[string]int{"a": 1}, false},
		{"map[string]any", map[string]any{"a": 1}, map[string]int{"a": 1}, false},
		{"map[any]int", map[any]int{"a": 1}, map[string]int{"a": 1}, false},
		{"map[any]any", map[any]any{"a": 1}, map[string]int{"a": 1}, false},
		{"string JSON", `{"a":1}`, map[string]int{"a": 1}, false},
		{"nil", nil, nil, true},
		{"unsupported", 42, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringMapIntE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestToStringMapNumberE_AllTypes(t *testing.T) {
	// Test all numeric map types
	if got, err := ToStringMapInt8E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapInt8E = %v, %v", got, err)
	}
	if got, err := ToStringMapInt16E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapInt16E = %v, %v", got, err)
	}
	if got, err := ToStringMapInt32E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapInt32E = %v, %v", got, err)
	}
	if got, err := ToStringMapInt64E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapInt64E = %v, %v", got, err)
	}
	if got, err := ToStringMapUintE(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapUintE = %v, %v", got, err)
	}
	if got, err := ToStringMapUint8E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapUint8E = %v, %v", got, err)
	}
	if got, err := ToStringMapUint16E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapUint16E = %v, %v", got, err)
	}
	if got, err := ToStringMapUint32E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapUint32E = %v, %v", got, err)
	}
	if got, err := ToStringMapUint64E(map[string]any{"a": 1}); err != nil || got["a"] != 1 {
		t.Errorf("ToStringMapUint64E = %v, %v", got, err)
	}
}

func TestToStringMapIntE_Unexported(t *testing.T) {
	// Test all branches of toStringMapIntE
	if got, err := toStringMapIntE(map[string]int{"a": 1}, ToInt, ToIntE); err != nil || got["a"] != 1 {
		t.Errorf("toStringMapIntE(map[string]int) = %v, %v", got, err)
	}
	if got, err := toStringMapIntE(map[string]any{"a": 1}, ToInt, ToIntE); err != nil || got["a"] != 1 {
		t.Errorf("toStringMapIntE(map[string]any) = %v, %v", got, err)
	}
	if got, err := toStringMapIntE(map[any]int{"a": 1}, ToInt, ToIntE); err != nil || got["a"] != 1 {
		t.Errorf("toStringMapIntE(map[any]int) = %v, %v", got, err)
	}
	if got, err := toStringMapIntE(map[any]any{"a": 1}, ToInt, ToIntE); err != nil || got["a"] != 1 {
		t.Errorf("toStringMapIntE(map[any]any) = %v, %v", got, err)
	}
	if got, err := toStringMapIntE(`{"a":1}`, ToInt, ToIntE); err != nil || got["a"] != 1 {
		t.Errorf("toStringMapIntE(string JSON) = %v, %v", got, err)
	}
	if _, err := toStringMapIntE(nil, ToInt, ToIntE); err == nil {
		t.Error("toStringMapIntE(nil) expected error")
	}
	if _, err := toStringMapIntE(42, ToInt, ToIntE); err == nil {
		t.Error("toStringMapIntE(42) expected error")
	}

	// Reflect map path
	type MyVal int
	m := map[string]MyVal{"a": 1}
	got, err := toStringMapIntE(m, ToInt, ToIntE)
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != 1 {
		t.Errorf("got %v", got)
	}

	// Reflect map error path
	type MyVal2 string
	m2 := map[string]MyVal2{"a": "invalid"}
	_, err = toStringMapIntE(m2, ToInt, ToIntE)
	if err == nil {
		t.Error("expected error")
	}
}

func TestToStringMapNumberE_ReflectMap(t *testing.T) {
	// Test with a map type not directly handled by switch (string keys, named value type)
	type MyVal int
	m := map[string]MyVal{"a": 1}
	got, err := ToStringMapIntE(m)
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != 1 {
		t.Errorf("got %v", got)
	}

	// Test error in value conversion
	type MyVal2 string
	m2 := map[string]MyVal2{"a": "invalid"}
	_, err = ToStringMapIntE(m2)
	if err == nil {
		t.Error("expected error")
	}
}

func TestToStrMapE(t *testing.T) {
	// Test each type
	if _, err := ToStrMapE[uint8](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[uint8] error = %v", err)
	}
	if _, err := ToStrMapE[uint16](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[uint16] error = %v", err)
	}
	if _, err := ToStrMapE[uint32](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[uint32] error = %v", err)
	}
	if _, err := ToStrMapE[uint64](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[uint64] error = %v", err)
	}
	if _, err := ToStrMapE[uint](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[uint] error = %v", err)
	}
	if _, err := ToStrMapE[int8](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[int8] error = %v", err)
	}
	if _, err := ToStrMapE[int16](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[int16] error = %v", err)
	}
	if _, err := ToStrMapE[int32](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[int32] error = %v", err)
	}
	if _, err := ToStrMapE[int64](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[int64] error = %v", err)
	}
	if _, err := ToStrMapE[int](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[int] error = %v", err)
	}
	if _, err := ToStrMapE[float32](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[float32] error = %v", err)
	}
	if _, err := ToStrMapE[float64](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[float64] error = %v", err)
	}
	if _, err := ToStrMapE[string](map[string]any{"a": "1"}); err != nil {
		t.Errorf("ToStrMapE[string] error = %v", err)
	}
	if _, err := ToStrMapE[bool](map[string]any{"a": true}); err != nil {
		t.Errorf("ToStrMapE[bool] error = %v", err)
	}
	if _, err := ToStrMapE[time.Time](map[string]any{"a": "2023-01-15"}); err != nil {
		t.Errorf("ToStrMapE[time.Time] error = %v", err)
	}
	if _, err := ToStrMapE[time.Duration](map[string]any{"a": "1h"}); err != nil {
		t.Errorf("ToStrMapE[time.Duration] error = %v", err)
	}
	if _, err := ToStrMapE[map[string]any](map[string]any{"a": map[string]any{"b": 1}}); err != nil {
		t.Errorf("ToStrMapE[map[string]any] error = %v", err)
	}
	if _, err := ToStrMapE[any](map[string]any{"a": 1}); err != nil {
		t.Errorf("ToStrMapE[any] error = %v", err)
	}
	// Test error paths for ToStrMapE
	if _, err := ToStrMapE[uint8](42); err == nil {
		t.Error("ToStrMapE[uint8](42) expected error")
	}
	if _, err := ToStrMapE[uint16](42); err == nil {
		t.Error("ToStrMapE[uint16](42) expected error")
	}
	if _, err := ToStrMapE[uint32](42); err == nil {
		t.Error("ToStrMapE[uint32](42) expected error")
	}
	if _, err := ToStrMapE[uint64](42); err == nil {
		t.Error("ToStrMapE[uint64](42) expected error")
	}
	if _, err := ToStrMapE[uint](42); err == nil {
		t.Error("ToStrMapE[uint](42) expected error")
	}
	if _, err := ToStrMapE[int8](42); err == nil {
		t.Error("ToStrMapE[int8](42) expected error")
	}
	if _, err := ToStrMapE[int16](42); err == nil {
		t.Error("ToStrMapE[int16](42) expected error")
	}
	if _, err := ToStrMapE[int32](42); err == nil {
		t.Error("ToStrMapE[int32](42) expected error")
	}
	if _, err := ToStrMapE[int64](42); err == nil {
		t.Error("ToStrMapE[int64](42) expected error")
	}
	if _, err := ToStrMapE[int](42); err == nil {
		t.Error("ToStrMapE[int](42) expected error")
	}
	if _, err := ToStrMapE[float32](42); err == nil {
		t.Error("ToStrMapE[float32](42) expected error")
	}
	if _, err := ToStrMapE[float64](42); err == nil {
		t.Error("ToStrMapE[float64](42) expected error")
	}
	if _, err := ToStrMapE[string](42); err == nil {
		t.Error("ToStrMapE[string](42) expected error")
	}
	if _, err := ToStrMapE[bool](42); err == nil {
		t.Error("ToStrMapE[bool](42) expected error")
	}
	if _, err := ToStrMapE[time.Time](42); err == nil {
		t.Error("ToStrMapE[time.Time](42) expected error")
	}
	if _, err := ToStrMapE[time.Duration](42); err == nil {
		t.Error("ToStrMapE[time.Duration](42) expected error")
	}
	if _, err := ToStrMapE[map[string]any](42); err == nil {
		t.Error("ToStrMapE[map[string]any](42) expected error")
	}
	if _, err := ToStrMapE[any](42); err == nil {
		t.Error("ToStrMapE[any](42) expected error")
	}
}

func TestJsonStringToObject(t *testing.T) {
	var m map[string]any
	err := jsonStringToObject(`{"a":1}`, &m)
	if err != nil {
		t.Fatal(err)
	}
	if m["a"] != float64(1) {
		t.Errorf("got %v", m)
	}

	err = jsonStringToObject("invalid", &m)
	if err == nil {
		t.Error("expected error")
	}
}

func TestToStringMapSliceE(t *testing.T) {
	got, err := ToStringMapSliceE([]map[string]any{{"a": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("got %v", got)
	}

	_, err = ToStringMapSliceE("invalid")
	if err == nil {
		t.Error("expected error")
	}
}

func TestToMapE(t *testing.T) {
	// Test toMapE directly via exported wrappers
	// map[string]int is covered by ToStringMapIntE etc.
	// Let's test some edge cases via json
	got, err := ToStringMapIntE(`{"a":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != 1 {
		t.Errorf("got %v", got)
	}

	// Invalid JSON
	_, err = ToStringMapIntE("invalid")
	if err == nil {
		t.Error("expected error")
	}
}
