package confy

import (
	"testing"
)

type mockFlag struct {
	name    string
	val     string
	valType string
	changed bool
}

func (m mockFlag) HasChanged() bool { return m.changed }
func (m mockFlag) Name() string     { return m.name }
func (m mockFlag) ValueString() string { return m.val }
func (m mockFlag) ValueType() string   { return m.valType }

func TestBindFlagValue(t *testing.T) {
	Reset()
	flag := mockFlag{name: "port", val: "8080", valType: "int", changed: true}
	err := BindFlagValue("port", flag)
	if err != nil {
		t.Fatal(err)
	}

	val := GetInt("port")
	if val != 8080 {
		t.Fatalf("expected 8080, got %d", val)
	}
}

func TestBindFlagValueNotChanged(t *testing.T) {
	Reset()
	flag := mockFlag{name: "port", val: "8080", valType: "int", changed: false}
	err := BindFlagValue("port", flag)
	if err != nil {
		t.Fatal(err)
	}

	// flag not changed and no default -> should return nil
	val := GetConfy().find("port", false)
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}
