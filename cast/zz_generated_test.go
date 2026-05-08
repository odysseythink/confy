package cast

import (
	"reflect"
	"testing"
	"time"
)

func TestGeneratedToFunctions(t *testing.T) {
	if got := ToBool(true); got != true {
		t.Errorf("ToBool(true) = %v", got)
	}
	if got := ToBool("invalid"); got != false {
		t.Errorf("ToBool(\"invalid\") = %v", got)
	}

	if got := ToString("hello"); got != "hello" {
		t.Errorf("ToString(\"hello\") = %q", got)
	}
	if got := ToString(42); got != "42" {
		t.Errorf("ToString(42) = %q", got)
	}

	tm := time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)
	if got := ToTime(tm); !got.Equal(tm) {
		t.Errorf("ToTime = %v", got)
	}
	if got := ToTime("2023-01-15"); got.Year() != 2023 {
		t.Errorf("ToTime = %v", got)
	}

	if got := ToTimeInDefaultLocation("2023-01-15", time.UTC); got.Year() != 2023 {
		t.Errorf("ToTimeInDefaultLocation = %v", got)
	}

	if got := ToDuration(time.Hour); got != time.Hour {
		t.Errorf("ToDuration = %v", got)
	}
	if got := ToDuration("1h"); got != time.Hour {
		t.Errorf("ToDuration = %v", got)
	}

	if got := ToInt(42); got != 42 {
		t.Errorf("ToInt = %d", got)
	}
	if got := ToInt8(42); got != 42 {
		t.Errorf("ToInt8 = %d", got)
	}
	if got := ToInt16(42); got != 42 {
		t.Errorf("ToInt16 = %d", got)
	}
	if got := ToInt32(42); got != 42 {
		t.Errorf("ToInt32 = %d", got)
	}
	if got := ToInt64(42); got != 42 {
		t.Errorf("ToInt64 = %d", got)
	}

	if got := ToUint(42); got != 42 {
		t.Errorf("ToUint = %d", got)
	}
	if got := ToUint8(42); got != 42 {
		t.Errorf("ToUint8 = %d", got)
	}
	if got := ToUint16(42); got != 42 {
		t.Errorf("ToUint16 = %d", got)
	}
	if got := ToUint32(42); got != 42 {
		t.Errorf("ToUint32 = %d", got)
	}
	if got := ToUint64(42); got != 42 {
		t.Errorf("ToUint64 = %d", got)
	}

	if got := ToFloat32(float32(3.14)); got != float32(3.14) {
		t.Errorf("ToFloat32 = %v", got)
	}
	if got := ToFloat64(3.14); got != 3.14 {
		t.Errorf("ToFloat64 = %v", got)
	}

	if got := ToStringMapString(map[string]string{"a": "1"}); !reflect.DeepEqual(got, map[string]string{"a": "1"}) {
		t.Errorf("ToStringMapString = %v", got)
	}
	if got := ToStringMapStringSlice(map[string][]string{"a": {"1"}}); !reflect.DeepEqual(got, map[string][]string{"a": {"1"}}) {
		t.Errorf("ToStringMapStringSlice = %v", got)
	}
	if got := ToStringMapBool(map[string]bool{"a": true}); !reflect.DeepEqual(got, map[string]bool{"a": true}) {
		t.Errorf("ToStringMapBool = %v", got)
	}
	if got := ToStringMapInt(map[string]int{"a": 1}); !reflect.DeepEqual(got, map[string]int{"a": 1}) {
		t.Errorf("ToStringMapInt = %v", got)
	}
	if got := ToStringMapInt64(map[string]int64{"a": 1}); !reflect.DeepEqual(got, map[string]int64{"a": 1}) {
		t.Errorf("ToStringMapInt64 = %v", got)
	}
	if got := ToStringMap(map[string]any{"a": 1}); !reflect.DeepEqual(got, map[string]any{"a": 1}) {
		t.Errorf("ToStringMap = %v", got)
	}
	if got := ToSlice([]any{1}); !reflect.DeepEqual(got, []any{1}) {
		t.Errorf("ToSlice = %v", got)
	}
}

func TestGeneratedSliceFunctions(t *testing.T) {
	if got := ToBoolSlice([]bool{true}); !reflect.DeepEqual(got, []bool{true}) {
		t.Errorf("ToBoolSlice = %v", got)
	}
	if got := ToDurationSlice([]time.Duration{time.Hour}); !reflect.DeepEqual(got, []time.Duration{time.Hour}) {
		t.Errorf("ToDurationSlice = %v", got)
	}
	if got := ToIntSlice([]int{1}); !reflect.DeepEqual(got, []int{1}) {
		t.Errorf("ToIntSlice = %v", got)
	}
	if got := ToInt8Slice([]int8{1}); !reflect.DeepEqual(got, []int8{1}) {
		t.Errorf("ToInt8Slice = %v", got)
	}
	if got := ToInt16Slice([]int16{1}); !reflect.DeepEqual(got, []int16{1}) {
		t.Errorf("ToInt16Slice = %v", got)
	}
	if got := ToInt32Slice([]int32{1}); !reflect.DeepEqual(got, []int32{1}) {
		t.Errorf("ToInt32Slice = %v", got)
	}
	if got := ToInt64Slice([]int64{1}); !reflect.DeepEqual(got, []int64{1}) {
		t.Errorf("ToInt64Slice = %v", got)
	}
	if got := ToUintSlice([]uint{1}); !reflect.DeepEqual(got, []uint{1}) {
		t.Errorf("ToUintSlice = %v", got)
	}
	if got := ToUint8Slice([]uint8{1}); !reflect.DeepEqual(got, []uint8{1}) {
		t.Errorf("ToUint8Slice = %v", got)
	}
	if got := ToUint16Slice([]uint16{1}); !reflect.DeepEqual(got, []uint16{1}) {
		t.Errorf("ToUint16Slice = %v", got)
	}
	if got := ToUint32Slice([]uint32{1}); !reflect.DeepEqual(got, []uint32{1}) {
		t.Errorf("ToUint32Slice = %v", got)
	}
	if got := ToUint64Slice([]uint64{1}); !reflect.DeepEqual(got, []uint64{1}) {
		t.Errorf("ToUint64Slice = %v", got)
	}
	if got := ToFloat32Slice([]float32{1.1}); !reflect.DeepEqual(got, []float32{1.1}) {
		t.Errorf("ToFloat32Slice = %v", got)
	}
	if got := ToFloat64Slice([]float64{1.1}); !reflect.DeepEqual(got, []float64{1.1}) {
		t.Errorf("ToFloat64Slice = %v", got)
	}
}
