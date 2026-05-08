package pflag

import (
	"fmt"

	"github.com/odysseythink/confy"
	"github.com/spf13/pflag"
)

func BindPFlag(v *confy.Confy, key string, flag *pflag.Flag) error {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	return v.BindFlagValue(key, pflagValue{flag})
}

func BindPFlags(v *confy.Confy, flags *pflag.FlagSet) error {
	return v.BindFlagValues(pflagValueSet{flags})
}

type pflagValueSet struct {
	flags *pflag.FlagSet
}

func (p pflagValueSet) VisitAll(fn func(confy.FlagValue)) {
	p.flags.VisitAll(func(flag *pflag.Flag) {
		fn(pflagValue{flag})
	})
}

type pflagValue struct {
	flag *pflag.Flag
}

func (p pflagValue) HasChanged() bool     { return p.flag.Changed }
func (p pflagValue) Name() string         { return p.flag.Name }
func (p pflagValue) ValueString() string  { return p.flag.Value.String() }
func (p pflagValue) ValueType() string    { return p.flag.Value.Type() }
