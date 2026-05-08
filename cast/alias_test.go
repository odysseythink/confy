package cast

import (
	"reflect"
	"testing"
)

func TestResolveAlias(t *testing.T) {
	type MyInt int
	type MyString string
	type MyFloat float64
	type MyBool bool
	type MyStruct struct{ A int }
	type MyUint uint16
	type MyInt8 int8
	type MyInt16 int16
	type MyInt32 int32
	type MyInt64 int64
	type MyUint8 uint8
	type MyUint16 uint16
	type MyUint32 uint32
	type MyUint64 uint64
	type MyFloat32 float32

	tests := []struct {
		name     string
		input    any
		wantVal  any
		wantBool bool
	}{
		{"nil", nil, nil, false},
		{"int", 42, 42, false},
		{"string", "hello", "hello", false},
		{"bool", true, true, false},
		{"float64", 3.14, 3.14, false},
		{"MyInt", MyInt(42), 42, true},
		{"MyString", MyString("hello"), "hello", true},
		{"MyFloat", MyFloat(3.14), 3.14, true},
		{"MyBool", MyBool(true), true, true},
		{"MyUint", MyUint(42), uint(42), true},
		{"MyInt8", MyInt8(42), int8(42), true},
		{"MyInt16", MyInt16(42), int16(42), true},
		{"MyInt32", MyInt32(42), int32(42), true},
		{"MyInt64", MyInt64(42), int64(42), true},
		{"MyUint8", MyUint8(42), uint8(42), true},
		{"MyUint16", MyUint16(42), uint16(42), true},
		{"MyUint32", MyUint32(42), uint32(42), true},
		{"MyUint64", MyUint64(42), uint64(42), true},
		{"MyFloat32", MyFloat32(3.14), float32(3.14), true},
		{"MyStruct", MyStruct{A: 1}, MyStruct{A: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotBool := resolveAlias(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("resolveAlias() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("resolveAlias() val = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
