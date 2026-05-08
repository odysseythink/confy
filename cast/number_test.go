package cast

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

// Test types for float64Provider / float64EProvider interfaces.
type testFloat64Provider struct {
	val float64
}

func (t testFloat64Provider) Float64() float64 {
	return t.val
}

type testFloat64EProvider struct {
	val float64
	err error
}

func (t testFloat64EProvider) Float64() (float64, error) {
	return t.val, t.err
}

func TestToNumberE(t *testing.T) {
	if got, err := ToNumberE[int](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[int](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[int8](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[int8](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[int16](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[int16](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[int32](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[int32](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[int64](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[int64](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[uint](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[uint](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[uint8](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[uint8](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[uint16](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[uint16](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[uint32](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[uint32](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[uint64](42); err != nil || got != 42 {
		t.Errorf("ToNumberE[uint64](42) = %v, %v", got, err)
	}
	if got, err := ToNumberE[float32](3.14); err != nil || math.Abs(float64(got)-3.14) > 0.01 {
		t.Errorf("ToNumberE[float32](3.14) = %v, %v", got, err)
	}
	if got, err := ToNumberE[float64](3.14); err != nil || math.Abs(got-3.14) > 0.01 {
		t.Errorf("ToNumberE[float64](3.14) = %v, %v", got, err)
	}
}

func TestToNumber(t *testing.T) {
	if got := ToNumber[int](42); got != 42 {
		t.Errorf("ToNumber[int](42) = %v", got)
	}
}

func TestToFloat64E(t *testing.T) {
	tests := []struct {
		input    any
		expected float64
		wantErr  bool
	}{
		{float64(3.14), 3.14, false},
		{float32(3.14), 3.14, false},
		{int(42), 42, false},
		{int8(42), 42, false},
		{int16(42), 42, false},
		{int32(42), 42, false},
		{int64(42), 42, false},
		{uint(42), 42, false},
		{uint8(42), 42, false},
		{uint16(42), 42, false},
		{uint32(42), 42, false},
		{uint64(42), 42, false},
		{true, 1, false},
		{false, 0, false},
		{nil, 0, false},
		{time.Weekday(1), 1, false},
		{time.Month(1), 1, false},
		{"3.14", 3.14, false},
		{"", 0, false},
		{"invalid", 0, true},
		{json.Number("3.14"), 3.14, false},
		{json.Number(""), 0, false},
		{json.Number("invalid"), 0, true},
		{testFloat64EProvider{val: 3.14}, 3.14, false},
		{testFloat64Provider{val: 3.14}, 3.14, false},
	}

	for _, tt := range tests {
		got, err := ToFloat64E(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToFloat64E(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && math.Abs(got-tt.expected) > 0.01 {
			t.Errorf("ToFloat64E(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}

	// Test alias
	type MyFloat float64
	if got, err := ToFloat64E(MyFloat(3.14)); err != nil || math.Abs(got-3.14) > 0.01 {
		t.Errorf("ToFloat64E(MyFloat(3.14)) = %v, %v", got, err)
	}

	// Test unsupported
	if _, err := ToFloat64E(struct{}{}); err == nil {
		t.Error("expected error")
	}

	// Test float64EProvider error
	_, err := ToFloat64E(testFloat64EProvider{val: 0, err: errors.New("fail")})
	if err == nil {
		t.Error("expected error for float64EProvider with error")
	}

	// Test float64Provider for non-float64 target (int)
	_, err = ToIntE(testFloat64Provider{val: 3.14})
	if err == nil {
		t.Error("expected error when using float64Provider for int")
	}

	// Test float64EProvider for non-float64 target (int)
	_, err = ToIntE(testFloat64EProvider{val: 3.14})
	if err == nil {
		t.Error("expected error when using float64EProvider for int")
	}
}

func TestToFloat32E(t *testing.T) {
	got, err := ToFloat32E(float32(3.14))
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(got)-3.14) > 0.01 {
		t.Errorf("got %v", got)
	}
}

func TestToInt64E(t *testing.T) {
	tests := []struct {
		input    any
		expected int64
		wantErr  bool
	}{
		{int64(42), 42, false},
		{int(42), 42, false},
		{int8(42), 42, false},
		{int16(42), 42, false},
		{int32(42), 42, false},
		{uint(42), 42, false},
		{uint8(42), 42, false},
		{uint16(42), 42, false},
		{uint32(42), 42, false},
		{uint64(42), 42, false},
		{float32(42), 42, false},
		{float64(42), 42, false},
		{true, 1, false},
		{false, 0, false},
		{nil, 0, false},
		{time.Weekday(1), 1, false},
		{time.Month(1), 1, false},
		{"42", 42, false},
		{"3.14", 3, false}, // trimDecimal strips decimal
		{"", 0, false},
		{"invalid", 0, true},
		{json.Number("42"), 42, false},
		{json.Number("3.14"), 3, false},
		{json.Number(""), 0, false},
		{json.Number("invalid"), 0, true},
	}

	for _, tt := range tests {
		got, err := ToInt64E(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToInt64E(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ToInt64E(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}

	// Test alias
	type MyInt int64
	if got, err := ToInt64E(MyInt(42)); err != nil || got != 42 {
		t.Errorf("ToInt64E(MyInt(42)) = %v, %v", got, err)
	}

	// Test unsupported
	if _, err := ToInt64E(struct{}{}); err == nil {
		t.Error("expected error")
	}
}

func TestToInt32E(t *testing.T) {
	got, err := ToInt32E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToInt16E(t *testing.T) {
	got, err := ToInt16E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToInt8E(t *testing.T) {
	got, err := ToInt8E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToIntE(t *testing.T) {
	got, err := ToIntE(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToUintE(t *testing.T) {
	tests := []struct {
		input    any
		expected uint
		wantErr  bool
	}{
		{uint(42), 42, false},
		{int(42), 42, false},
		{int8(42), 42, false},
		{int16(42), 42, false},
		{int32(42), 42, false},
		{int64(42), 42, false},
		{float32(42), 42, false},
		{float64(42), 42, false},
		{true, 1, false},
		{false, 0, false},
		{nil, 0, false},
		{time.Weekday(1), 1, false},
		{time.Month(1), 1, false},
		{"42", 42, false},
		{"", 0, false},
		{"invalid", 0, true},
		{json.Number("42"), 42, false},
	}

	for _, tt := range tests {
		got, err := ToUintE(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToUintE(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && got != tt.expected {
			t.Errorf("ToUintE(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}

	// Negative values
	if _, err := ToUintE(-1); err == nil {
		t.Error("ToUintE(-1) expected error")
	}
	if _, err := ToUintE(int8(-1)); err == nil {
		t.Error("ToUintE(int8(-1)) expected error")
	}
	if _, err := ToUintE(int16(-1)); err == nil {
		t.Error("ToUintE(int16(-1)) expected error")
	}
	if _, err := ToUintE(int32(-1)); err == nil {
		t.Error("ToUintE(int32(-1)) expected error")
	}
	if _, err := ToUintE(int64(-1)); err == nil {
		t.Error("ToUintE(int64(-1)) expected error")
	}
	if _, err := ToUintE(float32(-1)); err == nil {
		t.Error("ToUintE(float32(-1)) expected error")
	}
	if _, err := ToUintE(float64(-1)); err == nil {
		t.Error("ToUintE(float64(-1)) expected error")
	}
	if _, err := ToUintE("-1"); err == nil {
		t.Error("ToUintE(\"-1\") expected error")
	}
	if _, err := ToUintE(json.Number("-1")); err == nil {
		t.Error("ToUintE(json.Number(\"-1\")) expected error")
	}

	// Test alias
	type MyUint uint
	if got, err := ToUintE(MyUint(42)); err != nil || got != 42 {
		t.Errorf("ToUintE(MyUint(42)) = %v, %v", got, err)
	}
}

func TestToUint64E(t *testing.T) {
	got, err := ToUint64E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToUint32E(t *testing.T) {
	got, err := ToUint32E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToUint16E(t *testing.T) {
	got, err := ToUint16E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToUint8E(t *testing.T) {
	got, err := ToUint8E(42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToUnsignedNumber_CrossType(t *testing.T) {
	// Test cross-type conversions for unsigned numbers to hit case uint/uint8/etc.
	if got, err := ToUint16E(uint(42)); err != nil || got != 42 {
		t.Errorf("ToUint16E(uint(42)) = %v, %v", got, err)
	}
	if got, err := ToUint16E(uint8(42)); err != nil || got != 42 {
		t.Errorf("ToUint16E(uint8(42)) = %v, %v", got, err)
	}
	if got, err := ToUint32E(uint16(42)); err != nil || got != 42 {
		t.Errorf("ToUint32E(uint16(42)) = %v, %v", got, err)
	}
	if got, err := ToUint64E(uint32(42)); err != nil || got != 42 {
		t.Errorf("ToUint64E(uint32(42)) = %v, %v", got, err)
	}
	if got, err := ToUintE(uint64(42)); err != nil || got != 42 {
		t.Errorf("ToUintE(uint64(42)) = %v, %v", got, err)
	}
	if got, err := ToUint8E(uint(42)); err != nil || got != 42 {
		t.Errorf("ToUint8E(uint(42)) = %v, %v", got, err)
	}
}

func TestToUnsignedNumberE_DefaultError(t *testing.T) {
	// Test default error path in toUnsignedNumberE
	_, err := ToUintE(struct{}{})
	if err == nil {
		t.Error("ToUintE(struct{}{}) expected error")
	}
}

func TestToUintE_FloatProviderNegative(t *testing.T) {
	// float64Provider returning negative for unsigned target
	_, err := ToUintE(testFloat64Provider{val: -1.5})
	if err == nil {
		t.Error("expected error for negative float64Provider")
	}

	// float64EProvider returning negative for unsigned target
	_, err = ToUintE(testFloat64EProvider{val: -1.5})
	if err == nil {
		t.Error("expected error for negative float64EProvider")
	}

	// float64EProvider error for unsigned target
	_, err = ToUintE(testFloat64EProvider{val: 0, err: errors.New("fail")})
	if err == nil {
		t.Error("expected error for float64EProvider with error")
	}
}

func TestTrimZeroDecimal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.0", "1"},
		{"1.00", "1"},
		{"1.000", "1"},
		{"1", "1"},
		{"1.5", "1.5"},
		{"0", "0"},
		{".0", ""},
		{"0.", "0."},
		{"10.0", "10"},
		{"", ""},
		{"1.10", "1.10"},
		{"1.010", "1.010"},
	}

	for _, tt := range tests {
		got := trimZeroDecimal(tt.input)
		if got != tt.expected {
			t.Errorf("trimZeroDecimal(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestTrimDecimal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.5", "1"},
		{"1", "1"},
		{"-1.5", "-1"},
		{"+1.5", "+1"},
		{"-.", "-0"},
		{"+.", "+0"},
		{".5", "0"},
		{"abc", "abc"},
		{"1.5.6", "1.5.6"},
		{"-1.0", "-1"},
		{"+1.0", "+1"},
		{"", ""},
		{"1.", "1"},
		{"-1.", "-1"},
	}

	for _, tt := range tests {
		got := trimDecimal(tt.input)
		if got != tt.expected {
			t.Errorf("trimDecimal(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseInt(t *testing.T) {
	got, err := parseInt[int]("42")
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}

	_, err = parseInt[int]("invalid")
	if err == nil {
		t.Error("expected error")
	}
}

func TestParseUint(t *testing.T) {
	got, err := parseUint[uint]("42")
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}

	_, err = parseUint[uint]("invalid")
	if err == nil {
		t.Error("expected error")
	}

	// With + prefix
	got, err = parseUint[uint]("+42")
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestParseFloat(t *testing.T) {
	got, err := parseFloat[float64]("3.14")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got-3.14) > 0.01 {
		t.Errorf("got %f", got)
	}

	got32, err := parseFloat[float32]("3.14")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(float64(got32)-3.14) > 0.01 {
		t.Errorf("got %f", got32)
	}
}

func TestParseNumber(t *testing.T) {
	// Test all number types through parseNumber
	if got, err := parseNumber[int]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[int] = %v, %v", got, err)
	}
	if got, err := parseNumber[int8]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[int8] = %v, %v", got, err)
	}
	if got, err := parseNumber[int16]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[int16] = %v, %v", got, err)
	}
	if got, err := parseNumber[int32]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[int32] = %v, %v", got, err)
	}
	if got, err := parseNumber[int64]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[int64] = %v, %v", got, err)
	}
	if got, err := parseNumber[uint]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[uint] = %v, %v", got, err)
	}
	if got, err := parseNumber[uint8]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[uint8] = %v, %v", got, err)
	}
	if got, err := parseNumber[uint16]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[uint16] = %v, %v", got, err)
	}
	if got, err := parseNumber[uint32]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[uint32] = %v, %v", got, err)
	}
	if got, err := parseNumber[uint64]("42"); err != nil || got != 42 {
		t.Errorf("parseNumber[uint64] = %v, %v", got, err)
	}
	if got, err := parseNumber[float32]("3.14"); err != nil || math.Abs(float64(got)-3.14) > 0.01 {
		t.Errorf("parseNumber[float32] = %v, %v", got, err)
	}
	if got, err := parseNumber[float64]("3.14"); err != nil || math.Abs(got-3.14) > 0.01 {
		t.Errorf("parseNumber[float64] = %v, %v", got, err)
	}
}

func TestToUnsignedNumber_WeekdayMonthNegative(t *testing.T) {
	// time.Weekday and time.Month can be negative (in theory)
	_, _, ok := toUnsignedNumber[uint](time.Weekday(-1))
	if ok {
		t.Error("toUnsignedNumber should fail for negative weekday")
	}
	_, _, ok = toUnsignedNumber[uint](time.Month(-1))
	if ok {
		t.Error("toUnsignedNumber should fail for negative month")
	}
}
