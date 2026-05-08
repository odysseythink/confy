package internal

import (
	"testing"
	"time"
)

func TestTimeFormat_HasTimezone(t *testing.T) {
	tests := []struct {
		format TimeFormat
		want   bool
	}{
		{TimeFormat{Format: "2006-01-02", Typ: TimeFormatNoTimezone}, false},
		{TimeFormat{Format: time.RFC3339, Typ: TimeFormatNamedTimezone}, false},
		{TimeFormat{Format: time.RFC1123Z, Typ: TimeFormatNumericTimezone}, true},
		{TimeFormat{Format: "2006-01-02 15:04:05.999999999 -0700 MST", Typ: TimeFormatNumericAndNamedTimezone}, true},
		{TimeFormat{Format: time.Kitchen, Typ: TimeFormatTimeOnly}, false},
	}
	for _, tt := range tests {
		if got := tt.format.HasTimezone(); got != tt.want {
			t.Errorf("HasTimezone() = %v, want %v", got, tt.want)
		}
	}
}

func TestParseDateWith(t *testing.T) {
	loc := time.UTC

	t.Run("valid RFC3339", func(t *testing.T) {
		d, err := ParseDateWith("2023-01-15T10:30:00Z", loc, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2023 || d.Month() != time.January || d.Day() != 15 {
			t.Errorf("unexpected date: %v", d)
		}
	})

	t.Run("valid date only", func(t *testing.T) {
		d, err := ParseDateWith("2023-01-15", loc, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2023 || d.Month() != time.January || d.Day() != 15 {
			t.Errorf("unexpected date: %v", d)
		}
	})

	t.Run("valid with named timezone", func(t *testing.T) {
		d, err := ParseDateWith("Sun, 15 Jan 2023 10:30:00 UTC", nil, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2023 {
			t.Errorf("unexpected date: %v", d)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		_, err := ParseDateWith("not-a-date", loc, TimeFormats)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("kitchen format", func(t *testing.T) {
		d, err := ParseDateWith("3:04PM", loc, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Hour() != 15 || d.Minute() != 4 {
			t.Errorf("unexpected time: %v", d)
		}
	})

	t.Run("nil location uses local", func(t *testing.T) {
		d, err := ParseDateWith("2023-01-15", nil, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Location() != time.Local {
			t.Errorf("expected Local, got %v", d.Location())
		}
	})

	t.Run("RFC1123Z format", func(t *testing.T) {
		d, err := ParseDateWith("Sun, 15 Jan 2023 10:30:00 +0000", loc, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2023 {
			t.Errorf("unexpected date: %v", d)
		}
	})

	t.Run("numeric timezone format", func(t *testing.T) {
		d, err := ParseDateWith("2023-01-15T10:30:00-0700", loc, TimeFormats)
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2023 {
			t.Errorf("unexpected date: %v", d)
		}
	})
}

func TestTimeFormatType_String(t *testing.T) {
	tests := []struct {
		typ  TimeFormatType
		want string
	}{
		{TimeFormatNoTimezone, "TimeFormatNoTimezone"},
		{TimeFormatNamedTimezone, "TimeFormatNamedTimezone"},
		{TimeFormatNumericTimezone, "TimeFormatNumericTimezone"},
		{TimeFormatNumericAndNamedTimezone, "TimeFormatNumericAndNamedTimezone"},
		{TimeFormatTimeOnly, "TimeFormatTimeOnly"},
		{TimeFormatType(99), "TimeFormatType(99)"},
	}
	for _, tt := range tests {
		if got := tt.typ.String(); got != tt.want {
			t.Errorf("String() = %q, want %q", got, tt.want)
		}
	}
}
