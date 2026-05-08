package cast

import (
	"reflect"
	"testing"
	"time"
)

func TestToSliceE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []any
		wantErr  bool
	}{
		{"[]any", []any{1, 2}, []any{1, 2}, false},
		{"[]map[string]any", []map[string]any{{"a": 1}}, []any{map[string]any{"a": 1}}, false},
		{"unsupported", "hello", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToSliceE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}

	// indirect
	s := []any{1}
	if got, err := ToSliceE(&s); err != nil || !reflect.DeepEqual(got, []any{1}) {
		t.Errorf("ToSliceE(&s) = %v, %v", got, err)
	}
}

func TestToStringSliceE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
		wantErr  bool
	}{
		{"[]string", []string{"a", "b"}, []string{"a", "b"}, false},
		{"string", "a b c", []string{"a", "b", "c"}, false},
		{"int", 42, []string{"42"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringSliceE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}

	// Test error path via case any -> ToStringE failure
	_, err := ToStringSliceE(struct{}{})
	if err == nil {
		t.Error("ToStringSliceE(struct{}{}) expected error")
	}
}

func TestToSliceE_Generics(t *testing.T) {
	// Test toSliceE with various types
	got, err := toSliceE[int]([]int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}

	// Test with reflect slice (different element type)
	got, err = toSliceE[int]([]int8{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}

	// Test with array
	got, err = toSliceE[int]([2]int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}

	// Test error in element conversion
	_, err = toSliceE[int]([]any{"not-a-number"})
	if err == nil {
		t.Error("expected error")
	}

	// Test nil
	_, err = toSliceE[int](nil)
	if err == nil {
		t.Error("expected error")
	}

	// Test unsupported
	_, err = toSliceE[int]("hello")
	if err == nil {
		t.Error("expected error")
	}

	// Test direct []T match
	got, err = toSliceE[int]([]int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToBoolSliceE(t *testing.T) {
	got, err := ToBoolSliceE([]bool{true, false})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []bool{true, false}) {
		t.Errorf("got %v", got)
	}

	// Via reflect
	got, err = ToBoolSliceE([]any{true, false})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []bool{true, false}) {
		t.Errorf("got %v", got)
	}

	_, err = ToBoolSliceE([]any{"invalid"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestToDurationSliceE(t *testing.T) {
	got, err := ToDurationSliceE([]time.Duration{time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []time.Duration{time.Hour}) {
		t.Errorf("got %v", got)
	}

	_, err = ToDurationSliceE([]any{"invalid"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestToIntSliceE(t *testing.T) {
	got, err := ToIntSliceE([]int{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Errorf("got %v", got)
	}

	_, err = ToIntSliceE([]any{"invalid"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestToInt8SliceE(t *testing.T) {
	got, err := ToInt8SliceE([]int8{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int8{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToInt16SliceE(t *testing.T) {
	got, err := ToInt16SliceE([]int16{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int16{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToInt32SliceE(t *testing.T) {
	got, err := ToInt32SliceE([]int32{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int32{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToInt64SliceE(t *testing.T) {
	got, err := ToInt64SliceE([]int64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []int64{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToUintSliceE(t *testing.T) {
	got, err := ToUintSliceE([]uint{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []uint{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToUint8SliceE(t *testing.T) {
	got, err := ToUint8SliceE([]uint8{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []uint8{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToUint16SliceE(t *testing.T) {
	got, err := ToUint16SliceE([]uint16{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []uint16{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToUint32SliceE(t *testing.T) {
	got, err := ToUint32SliceE([]uint32{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []uint32{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToUint64SliceE(t *testing.T) {
	got, err := ToUint64SliceE([]uint64{1, 2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []uint64{1, 2}) {
		t.Errorf("got %v", got)
	}
}

func TestToFloat32SliceE(t *testing.T) {
	got, err := ToFloat32SliceE([]float32{1.1, 2.2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []float32{1.1, 2.2}) {
		t.Errorf("got %v", got)
	}
}

func TestToFloat64SliceE(t *testing.T) {
	got, err := ToFloat64SliceE([]float64{1.1, 2.2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []float64{1.1, 2.2}) {
		t.Errorf("got %v", got)
	}
}
