package cast

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestToE_String(t *testing.T) {
	got, err := ToE[string]("hello")
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestToE_Bool(t *testing.T) {
	got, err := ToE[bool](true)
	if err != nil {
		t.Fatal(err)
	}
	if got != true {
		t.Errorf("got %v", got)
	}
}

func TestToE_Int(t *testing.T) {
	got, err := ToE[int](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Int8(t *testing.T) {
	got, err := ToE[int8](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Int16(t *testing.T) {
	got, err := ToE[int16](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Int32(t *testing.T) {
	got, err := ToE[int32](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Int64(t *testing.T) {
	got, err := ToE[int64](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Uint(t *testing.T) {
	got, err := ToE[uint](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Uint8(t *testing.T) {
	got, err := ToE[uint8](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Uint16(t *testing.T) {
	got, err := ToE[uint16](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Uint32(t *testing.T) {
	got, err := ToE[uint32](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Uint64(t *testing.T) {
	got, err := ToE[uint64](42)
	if err != nil {
		t.Fatal(err)
	}
	if got != 42 {
		t.Errorf("got %d", got)
	}
}

func TestToE_Float32(t *testing.T) {
	got, err := ToE[float32](3.14)
	if err != nil {
		t.Fatal(err)
	}
	if got != float32(3.14) {
		t.Errorf("got %f", got)
	}
}

func TestToE_Float64(t *testing.T) {
	got, err := ToE[float64](3.14)
	if err != nil {
		t.Fatal(err)
	}
	if got != 3.14 {
		t.Errorf("got %f", got)
	}
}

func TestToE_Time(t *testing.T) {
	tm := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	got, err := ToE[time.Time](tm)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(tm) {
		t.Errorf("got %v", got)
	}
}

func TestToE_Duration(t *testing.T) {
	got, err := ToE[time.Duration](time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if got != time.Hour {
		t.Errorf("got %v", got)
	}
}

func TestToE_MapStringAny(t *testing.T) {
	m := map[string]any{"a": 1}
	got, err := ToE[map[string]any](m)
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != 1 {
		t.Errorf("got %v", got)
	}
}

func TestToE_SliceString(t *testing.T) {
	s := []string{"a", "b"}
	got, err := ToE[[]string](s)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "a" {
		t.Errorf("got %v", got)
	}
}

func TestToE_SliceMapStringAny(t *testing.T) {
	s := []map[string]any{{"a": 1}}
	got, err := ToE[[]map[string]any](s)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("got %v", got)
	}
}

func TestToE_SliceAny(t *testing.T) {
	s := []any{1, 2}
	got, err := ToE[[]any](s)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

func TestToE_SliceInt(t *testing.T) {
	s := []int{1, 2}
	got, err := ToE[[]int](s)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

func TestToE_Error(t *testing.T) {
	_, err := ToE[int](struct{}{})
	if err == nil {
		t.Error("expected error")
	}
}

func TestTo(t *testing.T) {
	if got := To[int](42); got != 42 {
		t.Errorf("To[int](42) = %d", got)
	}
	if got := To[string]("hello"); got != "hello" {
		t.Errorf("To[string](\"hello\") = %q", got)
	}
}

func TestMust(t *testing.T) {
	if got := Must[int](42, nil); got != 42 {
		t.Errorf("Must[int](42, nil) = %d", got)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	Must[int](0, errors.New("fail"))
}

func TestToE_Indirect(t *testing.T) {
	s := "hello"
	got, err := ToE[string](&s)
	if err != nil {
		t.Fatal(err)
	}
	if got != "hello" {
		t.Errorf("got %q", got)
	}
}

func TestToE_SliceReflect(t *testing.T) {
	// Test conversion via reflect for slice types
	s := []int8{1, 2}
	got, err := ToE[[]int](s)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}
}
