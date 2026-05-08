package cast

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestToTimeE(t *testing.T) {
	tm := time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    any
		expected time.Time
		wantErr  bool
	}{
		{"time.Time", tm, tm, false},
		{"string RFC3339", "2023-01-15T10:30:00Z", tm, false},
		{"string date only", "2023-01-15", time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC), false},
		{"json.Number", json.Number("1673779800"), time.Unix(1673779800, 0), false},
		{"json.Number with decimal", json.Number("1673779800.0"), time.Unix(1673779800, 0), false},
		{"json.Number invalid", json.Number("invalid"), time.Time{}, true},
		{"int", 1673779800, time.Unix(1673779800, 0), false},
		{"int32", int32(1673779800), time.Unix(1673779800, 0), false},
		{"int64", int64(1673779800), time.Unix(1673779800, 0), false},
		{"uint", uint(1673779800), time.Unix(1673779800, 0), false},
		{"uint32", uint32(1673779800), time.Unix(1673779800, 0), false},
		{"uint64", uint64(1673779800), time.Unix(1673779800, 0), false},
		{"nil", nil, time.Time{}, false},
		{"unsupported", struct{}{}, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToTimeE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !got.Equal(tt.expected) {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}

	// indirect
	pt := tm
	if got, err := ToTimeE(&pt); err != nil || !got.Equal(tm) {
		t.Errorf("ToTimeE(&pt) = %v, %v", got, err)
	}
}

func TestToTimeInDefaultLocationE(t *testing.T) {
	loc := time.FixedZone("TEST", 3600)
	got, err := ToTimeInDefaultLocationE("2023-01-15", loc)
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != loc {
		t.Errorf("got location %v, want %v", got.Location(), loc)
	}

	// nil location falls back to local
	got, err = ToTimeInDefaultLocationE("2023-01-15", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != time.Local {
		t.Errorf("got location %v, want Local", got.Location())
	}
}

func TestToDurationE(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected time.Duration
		wantErr  bool
	}{
		{"duration", time.Hour, time.Hour, false},
		{"int", 3600, time.Duration(3600), false},
		{"int8", int8(60), time.Duration(60), false},
		{"int16", int16(60), time.Duration(60), false},
		{"int32", int32(60), time.Duration(60), false},
		{"int64", int64(60), time.Duration(60), false},
		{"uint", uint(60), time.Duration(60), false},
		{"uint8", uint8(60), time.Duration(60), false},
		{"uint16", uint16(60), time.Duration(60), false},
		{"uint32", uint32(60), time.Duration(60), false},
		{"uint64", uint64(60), time.Duration(60), false},
		{"float32", float32(60), time.Duration(60), false},
		{"float64", float64(60), time.Duration(60), false},
		{"string with unit", "1h", time.Hour, false},
		{"string with unit ns", "1000000000ns", time.Second, false},
		{"string with unit us", "1000000µs", time.Second, false},
		{"string with unit ms", "1000ms", time.Second, false},
		{"string without unit", "3600000000000", time.Hour, false},
		{"string invalid", "invalid", 0, true},
		{"nil", nil, 0, false},
		{"unsupported", struct{}{}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToDurationE(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}

	// Test alias
	type MyDuration int64
	if got, err := ToDurationE(MyDuration(60)); err != nil || got != 60 {
		t.Errorf("ToDurationE(MyDuration(60)) = %v, %v", got, err)
	}

	// Test indirect
	d := time.Hour
	if got, err := ToDurationE(&d); err != nil || got != time.Hour {
		t.Errorf("ToDurationE(&d) = %v, %v", got, err)
	}
}

func TestStringToDate(t *testing.T) {
	got, err := StringToDate("2023-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if got.Year() != 2023 {
		t.Errorf("got %v", got)
	}

	_, err = StringToDate("invalid")
	if err == nil {
		t.Error("expected error")
	}
}

func TestStringToDateInDefaultLocation(t *testing.T) {
	loc := time.FixedZone("TEST", 3600)
	got, err := StringToDateInDefaultLocation("2023-01-15", loc)
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != loc {
		t.Errorf("got location %v, want %v", got.Location(), loc)
	}
}

func TestToTimeSliceE(t *testing.T) {
	tm1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	tm2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	got, err := ToTimeSliceE([]time.Time{tm1, tm2})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []time.Time{tm1, tm2}) {
		t.Errorf("got %v", got)
	}

	// Via reflect
	got, err = ToTimeSliceE([]any{"2023-01-01", "2023-01-02"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}

	_, err = ToTimeSliceE([]any{"invalid"})
	if err == nil {
		t.Error("expected error")
	}
}
