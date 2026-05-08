package confy

import (
	"fmt"
	"strings"
)

// FlagValueSet is an interface that users can implement
type FlagValueSet interface {
	VisitAll(fn func(FlagValue))
}

// FlagValue is an interface that users can implement
type FlagValue interface {
	HasChanged() bool
	Name() string
	ValueString() string
	ValueType() string
}

func BindFlagValues(flags FlagValueSet) error { return v.BindFlagValues(flags) }
func (v *Confy) BindFlagValues(flags FlagValueSet) (err error) {
	flags.VisitAll(func(flag FlagValue) {
		if err = v.BindFlagValue(flag.Name(), flag); err != nil {
			return
		}
	})
	return nil
}

func BindFlagValue(key string, flag FlagValue) error { return v.BindFlagValue(key, flag) }
func (v *Confy) BindFlagValue(key string, flag FlagValue) error {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	v.pflags[strings.ToLower(key)] = flag
	return nil
}
