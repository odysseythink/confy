package cast

import (
	"reflect"
	"testing"
)

func TestIndirect(t *testing.T) {
	var nilPtr *int
	i := 42
	ptr := &i
	ptrptr := &ptr

	var iface any = ptr
	var nilIface any = nilPtr
	var intIface any = 42

	tests := []struct {
		name     string
		input    any
		wantVal  any
		wantBool bool
	}{
		{"nil", nil, nil, false},
		{"int", 42, 42, false},
		{"*int", ptr, 42, true},
		{"**int", ptrptr, 42, true},
		{"nil *int", nilPtr, nil, true},
		{"interface with *int", iface, 42, true},
		{"interface with nil *int", nilIface, nil, true},
		{"interface with int", intIface, 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotBool := indirect(tt.input)
			if gotBool != tt.wantBool {
				t.Errorf("indirect() bool = %v, want %v", gotBool, tt.wantBool)
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("indirect() val = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
