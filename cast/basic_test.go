package cast

import (
	"encoding/json"
	"errors"
	"html/template"
	"testing"
	"time"
)

func TestToBoolE(t *testing.T) {
	tests := []struct {
		input    any
		expected bool
		wantErr  bool
	}{
		{true, true, false},
		{false, false, false},
		{nil, false, false},
		{int(1), true, false},
		{int(0), false, false},
		{int8(1), true, false},
		{int8(0), false, false},
		{int16(1), true, false},
		{int16(0), false, false},
		{int32(1), true, false},
		{int32(0), false, false},
		{int64(1), true, false},
		{int64(0), false, false},
		{uint(1), true, false},
		{uint(0), false, false},
		{uint8(1), true, false},
		{uint8(0), false, false},
		{uint16(1), true, false},
		{uint16(0), false, false},
		{uint32(1), true, false},
		{uint32(0), false, false},
		{uint64(1), true, false},
		{uint64(0), false, false},
		{float32(1), true, false},
		{float32(0), false, false},
		{float64(1), true, false},
		{float64(0), false, false},
		{time.Duration(1), true, false},
		{time.Duration(0), false, false},
		{"true", true, false},
		{"false", false, false},
		{"1", true, false},
		{"0", false, false},
		{"True", true, false},
		{"FALSE", false, false},
		{"invalid", false, true},
		{json.Number("1"), true, false},
		{json.Number("0"), false, false},
		{json.Number("invalid"), false, true},
		{json.Number(""), false, false},
	}

	for _, tt := range tests {
		got, err := ToBoolE(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToBoolE(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("ToBoolE(%v) = %v, want %v", tt.input, got, tt.expected)
		}
	}

	// Test alias resolution
	type MyBool bool
	if got, err := ToBoolE(MyBool(true)); err != nil || got != true {
		t.Errorf("ToBoolE(MyBool(true)) = %v, %v", got, err)
	}

	// Test unsupported type
	if _, err := ToBoolE(struct{}{}); err == nil {
		t.Error("ToBoolE(struct{}{}) expected error")
	}

	// Test indirect
	b := true
	if got, err := ToBoolE(&b); err != nil || got != true {
		t.Errorf("ToBoolE(&b) = %v, %v", got, err)
	}
}

func TestToStringE(t *testing.T) {
	tests := []struct {
		input    any
		expected string
		wantErr  bool
	}{
		{"hello", "hello", false},
		{true, "true", false},
		{false, "false", false},
		{float64(3.14), "3.14", false},
		{float32(2.5), "2.5", false},
		{int(42), "42", false},
		{int8(1), "1", false},
		{int16(2), "2", false},
		{int32(3), "3", false},
		{int64(4), "4", false},
		{uint(5), "5", false},
		{uint8(6), "6", false},
		{uint16(7), "7", false},
		{uint32(8), "8", false},
		{uint64(9), "9", false},
		{json.Number("10"), "10", false},
		{[]byte("bytes"), "bytes", false},
		{template.HTML("html"), "html", false},
		{template.URL("url"), "url", false},
		{template.JS("js"), "js", false},
		{template.CSS("css"), "css", false},
		{template.HTMLAttr("attr"), "attr", false},
		{nil, "", false},
	}

	for _, tt := range tests {
		got, err := ToStringE(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ToStringE(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("ToStringE(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}

	// Test Stringer
	got, err := ToStringE(time.Hour)
	if err != nil {
		t.Errorf("ToStringE(time.Hour) error = %v", err)
	}
	if got != "1h0m0s" {
		t.Errorf("ToStringE(time.Hour) = %q, want %q", got, "1h0m0s")
	}

	// Test error
	got, err = ToStringE(errors.New("oops"))
	if err != nil {
		t.Errorf("ToStringE(errors.New) error = %v", err)
	}
	if got != "oops" {
		t.Errorf("ToStringE(errors.New) = %q, want %q", got, "oops")
	}

	// Test indirect
	s := "hello"
	if got, err := ToStringE(&s); err != nil || got != "hello" {
		t.Errorf("ToStringE(&s) = %v, %v", got, err)
	}

	// Test alias
	type MyString string
	if got, err := ToStringE(MyString("alias")); err != nil || got != "alias" {
		t.Errorf("ToStringE(MyString) = %v, %v", got, err)
	}

	// Test unsupported type
	if _, err := ToStringE(struct{}{}); err == nil {
		t.Error("ToStringE(struct{}{}) expected error")
	}
}

func TestToBool(t *testing.T) {
	if got := ToBool(true); got != true {
		t.Errorf("ToBool(true) = %v", got)
	}
	if got := ToBool("invalid"); got != false {
		t.Errorf("ToBool(\"invalid\") = %v", got)
	}
}

func TestToString(t *testing.T) {
	if got := ToString(42); got != "42" {
		t.Errorf("ToString(42) = %q", got)
	}
}
