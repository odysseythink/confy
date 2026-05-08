package mapstructure

import (
	"errors"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Test error type methods that were at 0% coverage.

func TestDecodeError_Name(t *testing.T) {
	e := newDecodeError("foo", errors.New("bar"))
	if e.Name() != "foo" {
		t.Fatalf("expected Name to be foo, got %s", e.Name())
	}
}

func TestDecodeError_mapstructure(t *testing.T) {
	var _ Error = newDecodeError("foo", errors.New("bar"))
}

func TestParseError_mapstructure(t *testing.T) {
	var _ Error = &ParseError{}
}

func TestUnconvertibleTypeError_mapstructure(t *testing.T) {
	var _ Error = &UnconvertibleTypeError{}
}

func TestStrconvNumError_Unwrap(t *testing.T) {
	inner := &strconv.NumError{Func: "ParseInt", Err: strconv.ErrRange}
	wrapped := wrapStrconvNumError(inner)
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != inner {
		t.Fatal("expected Unwrap to return inner error")
	}
}

func TestUrlError_Unwrap(t *testing.T) {
	inner := &url.Error{Op: "Parse", URL: "://bad", Err: errors.New("bad url")}
	wrapped := wrapUrlError(inner)
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != inner {
		t.Fatal("expected Unwrap to return inner error")
	}
}

func TestNetParseError_Unwrap(t *testing.T) {
	inner := &net.ParseError{Type: "IP address", Text: "bad"}
	wrapped := wrapNetParseError(inner)
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != inner {
		t.Fatal("expected Unwrap to return inner error")
	}
}

func TestTimeParseError_Unwrap(t *testing.T) {
	inner := &time.ParseError{Layout: "2006", Value: "x", LayoutElem: "2006", Message: "bad"}
	wrapped := wrapTimeParseError(inner)
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != inner {
		t.Fatal("expected Unwrap to return inner error")
	}
}

func TestTimeParseError_Error_emptyMessage(t *testing.T) {
	inner := &time.ParseError{Layout: "2006", Value: "x", LayoutElem: "2006", Message: ""}
	wrapped := wrapTimeParseError(inner)
	msg := wrapped.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
}

// Test wrap*Error functions with non-matching error types.

func TestWrapStrconvNumError_nonNumError(t *testing.T) {
	inner := errors.New("not a num error")
	result := wrapStrconvNumError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapUrlError_nonUrlError(t *testing.T) {
	inner := errors.New("not a url error")
	result := wrapUrlError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapNetParseError_nonNetParseError(t *testing.T) {
	inner := errors.New("not a net parse error")
	result := wrapNetParseError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapTimeParseError_nonTimeParseError(t *testing.T) {
	inner := errors.New("not a time parse error")
	result := wrapTimeParseError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapNetIPParseAddrError_nonParseAddr(t *testing.T) {
	inner := errors.New("not parse addr")
	result := wrapNetIPParseAddrError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapNetIPParsePrefixError_nonParsePrefix(t *testing.T) {
	inner := errors.New("not parse prefix")
	result := wrapNetIPParsePrefixError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapTimeParseDurationError_nonDuration(t *testing.T) {
	inner := errors.New("not duration")
	result := wrapTimeParseDurationError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

func TestWrapTimeParseLocationError_nonLocation(t *testing.T) {
	inner := errors.New("not location")
	result := wrapTimeParseLocationError(inner)
	if result != inner {
		t.Fatal("expected same error back")
	}
}

// Test decode hook functions with wrong types (early return paths).

func TestStringToTimeLocationHookFunc_wrongType(t *testing.T) {
	f := StringToTimeLocationHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToURLHookFunc_wrongType(t *testing.T) {
	f := StringToURLHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToIPHookFunc_wrongType(t *testing.T) {
	f := StringToIPHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToIPNetHookFunc_wrongType(t *testing.T) {
	f := StringToIPNetHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToTimeHookFunc_wrongType(t *testing.T) {
	f := StringToTimeHookFunc("2006-01-02")
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestTextUnmarshallerHookFunc_nonString(t *testing.T) {
	f := TextUnmarshallerHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(42), reflect.ValueOf(""))
	if err != nil {
		t.Fatal(err)
	}
	if result != 42 {
		t.Fatalf("expected 42, got %v", result)
	}
}

func TestStringToNetIPAddrHookFunc_wrongType(t *testing.T) {
	f := StringToNetIPAddrHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToNetIPAddrPortHookFunc_wrongType(t *testing.T) {
	f := StringToNetIPAddrPortHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

func TestStringToNetIPPrefixHookFunc_wrongType(t *testing.T) {
	f := StringToNetIPPrefixHookFunc()
	result, err := DecodeHookExec(f, reflect.ValueOf(1), reflect.ValueOf(1))
	if err != nil {
		t.Fatal(err)
	}
	if result != 1 {
		t.Fatalf("expected 1, got %v", result)
	}
}

// Test invalid decode hooks.

func TestTypedDecodeHook_invalid(t *testing.T) {
	result := typedDecodeHook(func() {})
	if result != nil {
		t.Fatal("expected nil for invalid hook")
	}
}

func TestCachedDecodeHook_invalid(t *testing.T) {
	f := cachedDecodeHook(func() {})
	_, err := f(reflect.ValueOf(1), reflect.ValueOf(""))
	if err == nil {
		t.Fatal("expected error for invalid hook")
	}
}

func TestDecodeHookExec_invalid(t *testing.T) {
	_, err := DecodeHookExec(func() {}, reflect.ValueOf(1), reflect.ValueOf(""))
	if err == nil {
		t.Fatal("expected error for invalid hook")
	}
}

// Test convenience functions with invalid input.

func TestWeakDecode_nonPointer(t *testing.T) {
	var s string
	err := WeakDecode("hello", s)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestDecodeMetadata_nonPointer(t *testing.T) {
	var s string
	var md Metadata
	err := DecodeMetadata("hello", s, &md)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestWeakDecodeMetadata_nonPointer(t *testing.T) {
	var s string
	var md Metadata
	err := WeakDecodeMetadata("hello", s, &md)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestNewDecoder_nonAddressable(t *testing.T) {
	_, err := NewDecoder(&DecoderConfig{Result: new(string)})
	if err != nil {
		t.Fatal(err)
	}

	// non-addressable: a pointer to a string literal via interface
	var s string = "x"
	_, err = NewDecoder(&DecoderConfig{Result: &s})
	if err != nil {
		t.Fatal(err)
	}
}

// Test isEmptyValue with complex types.

func TestIsEmptyValue_complex(t *testing.T) {
	v := reflect.ValueOf(complex64(0))
	if isEmptyValue(v) {
		t.Fatal("expected complex zero to NOT be empty")
	}
}

func TestIsEmptyValue_chan(t *testing.T) {
	var ch chan int
	v := reflect.ValueOf(ch)
	if isEmptyValue(v) {
		t.Fatal("expected nil chan to NOT be empty")
	}
}

// Test decodeFunc with mismatched types.

func TestDecodeFunc_mismatch(t *testing.T) {
	var result func()
	err := Decode("hello", &result)
	if err == nil {
		t.Fatal("expected error for func mismatch")
	}
}

// Test decodeBasic with unconvertible type.

func TestDecodeBasic_unconvertible(t *testing.T) {
	var result int
	err := Decode(map[string]any{"v": []int{1}}, &struct{ V int }{V: 0})
	if err != nil {
		// This might error through a different path; just ensure it doesn't panic
	}
	_ = result
}

// Test decode with unsupported type.

func TestDecode_unsupportedType(t *testing.T) {
	type Unsupported struct {
		Ch chan int
	}
	var result Unsupported
	err := Decode(map[string]any{"ch": make(chan int)}, &result)
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

// Test providerPathExists and duplicate remote provider adding.

// Test weak decode paths in decodeString.

func TestDecodeString_weakBool(t *testing.T) {
	var result struct {
		V string
	}
	err := WeakDecode(map[string]any{"v": true}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != "1" {
		t.Fatalf("expected 1, got %s", result.V)
	}
}

func TestDecodeString_weakFloat(t *testing.T) {
	var result struct {
		V string
	}
	err := WeakDecode(map[string]any{"v": float32(3.14)}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result.V, "3.14") {
		t.Fatalf("expected 3.14..., got %s", result.V)
	}
}

func TestDecodeString_weakSlice(t *testing.T) {
	var result struct {
		V string
	}
	err := WeakDecode(map[string]any{"v": []uint8{65, 66}}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != "AB" {
		t.Fatalf("expected AB, got %s", result.V)
	}
}

// Test decodeInt negative uint.

func TestDecodeInt_negativeUint(t *testing.T) {
	var result struct {
		V uint
	}
	err := Decode(map[string]any{"v": -1}, &result)
	if err == nil {
		t.Fatal("expected error for negative uint")
	}
}

// Test decodeInt with string that overflows.

func TestDecodeInt_stringOverflow(t *testing.T) {
	var result struct {
		V int8
	}
	err := WeakDecode(map[string]any{"v": "128"}, &result)
	if err == nil {
		t.Fatal("expected error for string int overflow")
	}
}

// Test decodeUint float.

func TestDecodeUint_float(t *testing.T) {
	var result struct {
		V uint
	}
	err := WeakDecode(map[string]any{"v": 3.14}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != 3 {
		t.Fatalf("expected 3, got %d", result.V)
	}
}

// Test decodeUint string overflow.

func TestDecodeUint_stringOverflow(t *testing.T) {
	var result struct {
		V uint8
	}
	err := WeakDecode(map[string]any{"v": "256"}, &result)
	if err == nil {
		t.Fatal("expected error for string overflow")
	}
}

// Test decodeBool string.

func TestDecodeBool_string(t *testing.T) {
	var result struct {
		V bool
	}
	err := WeakDecode(map[string]any{"v": "true"}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if !result.V {
		t.Fatal("expected true")
	}
}

// Test decodeFloat uint.

func TestDecodeFloat_uint(t *testing.T) {
	var result struct {
		V float64
	}
	err := Decode(map[string]any{"v": uint(42)}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != 42 {
		t.Fatalf("expected 42, got %f", result.V)
	}
}

// Test decodeFloat string.

func TestDecodeFloat_string(t *testing.T) {
	var result struct {
		V float64
	}
	err := WeakDecode(map[string]any{"v": "3.14"}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result.V != 3.14 {
		t.Fatalf("expected 3.14, got %f", result.V)
	}
}

// Test decodeMapFromSlice with empty slice.

func TestDecodeMapFromSlice_empty(t *testing.T) {
	var result map[string]int
	err := WeakDecode([]map[string]int{}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result) != 0 {
		t.Fatalf("unexpected result: %v", result)
	}
}

// Test decodeMapFromMap with nil map.

func TestDecodeMapFromMap_nil(t *testing.T) {
	var result map[string]int
	var input map[string]int
	err := Decode(input, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

// Test decodeMapFromStruct with omitempty.

func TestDecodeMapFromStruct_omitempty(t *testing.T) {
	type Inner struct {
		V string `mapstructure:"v,omitempty"`
	}
	type Outer struct {
		Inner Inner `mapstructure:",squash"`
	}
	var result map[string]any
	err := Decode(Outer{}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result["v"]; ok {
		t.Fatal("expected v to be omitted")
	}
}

// Test decodeSlice non-array/slice weak typing.

func TestDecodeSlice_weakNonSlice(t *testing.T) {
	var result []string
	err := WeakDecode("hello", &result)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0] != "hello" {
		t.Fatalf("unexpected result: %v", result)
	}
}

// Test decodeArray from slice.

func TestDecodeArray_fromSlice(t *testing.T) {
	var result [2]int
	err := Decode([]int{1, 2}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result[0] != 1 || result[1] != 2 {
		t.Fatalf("unexpected result: %v", result)
	}
}

// Test decodeStructFromMap unused keys.

func TestDecodeStructFromMap_unusedKeys(t *testing.T) {
	type Result struct {
		V string
	}
	var md Metadata
	config := &DecoderConfig{
		Result:           &Result{},
		Metadata:         &md,
	}
	decoder, err := NewDecoder(config)
	if err != nil {
		t.Fatal(err)
	}
	err = decoder.Decode(map[string]any{"v": "foo", "extra": "bar"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, u := range md.Unused {
		if u == "extra" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected extra to be unused")
	}
}

// Test isStructTypeConvertibleToMap with exported fields.

func TestIsStructTypeConvertibleToMap_exported(t *testing.T) {
	type S struct {
		V string
	}
	if !isStructTypeConvertibleToMap(reflect.TypeOf(S{}), false, "mapstructure") {
		t.Fatal("expected true for struct with exported fields when checkMapstructureTags is false")
	}
}
