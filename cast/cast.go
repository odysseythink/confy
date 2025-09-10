// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package cast provides easy and safe casting in Go.
package cast

import "time"

const errorMsg = "unable to cast %#v of type %T to %T"
const errorMsgWith = "unable to cast %#v of type %T to %T: %w"

// Basic is a type parameter constraint for functions accepting basic types.
//
// It represents the supported basic types this package can cast to.
type Basic interface {
	string | bool | Number | time.Time | time.Duration | map[string]any
}

// ToE casts any value to a [Basic] type.
func ToE[T Basic | []map[string]any | []string | []any | []int](i any) (T, error) {
	var t T

	var v any
	var err error

	switch any(t).(type) {
	case string:
		v, err = ToStringE(i)
	case bool:
		v, err = ToBoolE(i)
	case int:
		v, err = toNumberE(i, parseInt[int])
	case int8:
		v, err = toNumberE(i, parseInt[int8])
	case int16:
		v, err = toNumberE(i, parseInt[int16])
	case int32:
		v, err = toNumberE(i, parseInt[int32])
	case int64:
		v, err = toNumberE(i, parseInt[int64])
	case uint:
		v, err = toUnsignedNumberE(i, parseUint[uint])
	case uint8:
		v, err = toUnsignedNumberE(i, parseUint[uint8])
	case uint16:
		v, err = toUnsignedNumberE(i, parseUint[uint16])
	case uint32:
		v, err = toUnsignedNumberE(i, parseUint[uint32])
	case uint64:
		v, err = toUnsignedNumberE(i, parseUint[uint64])
	case float32:
		v, err = toNumberE(i, parseFloat[float32])
	case float64:
		v, err = toNumberE(i, parseFloat[float64])
	case time.Time:
		v, err = ToTimeE(i)
	case time.Duration:
		v, err = ToDurationE(i)
	case map[string]any:
		v, err = ToStringMapE(i)
	case []string:
		v, err = toSliceE[string](i)
	case []map[string]any:
		v, err = toSliceE[map[string]any](i)
	case []any:
		v, err = ToSliceE(i)
	case []int:
		v, err = toSliceE[int](i)
	}

	if err != nil {
		return t, err
	}

	return v.(T), nil
}

// Must is a helper that wraps a call to a cast function and panics if the error is non-nil.
func Must[T any](i any, err error) T {
	if err != nil {
		panic(err)
	}

	return i.(T)
}

// To casts any value to a [Basic] type.
func To[T Basic | []map[string]any | []string | []any | []int](i any) T {
	v, _ := ToE[T](i)

	return v
}
