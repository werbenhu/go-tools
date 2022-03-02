package kson

import (
	"fmt"
	"strconv"
	"strings"

	"git.aimore.com/golang/kson/fastfloat"
)

// Value represents any JSON value.
//
// Call Type in order to determine the actual type of the JSON value.
//
// Value cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Value struct {
	o Object
	a []*Value
	s string
	t Type
}

// MarshalTo appends marshaled v to dst and returns the result.
func (v *Value) MarshalTo(dst []byte) []byte {
	switch v.t {
	case typeRawString:
		dst = append(dst, '"')
		dst = append(dst, v.s...)
		dst = append(dst, '"')
		return dst
	case TypeObject:
		return v.o.MarshalTo(dst)
	case TypeArray:
		dst = append(dst, '[')
		for i, vv := range v.a {
			dst = vv.MarshalTo(dst)
			if i != len(v.a)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ']')
		return dst
	case TypeString:
		return escapeString(dst, v.s)
	case TypeNumber:
		return append(dst, v.s...)
	case TypeTrue:
		return append(dst, "true"...)
	case TypeFalse:
		return append(dst, "false"...)
	case TypeNull:
		return append(dst, "null"...)
	default:
		panic(fmt.Errorf("BUG: unexpected Value type: %d", v.t))
	}
}

// String returns string representation of the v.
//
// The function is for debugging purposes only. It isn't optimized for speed.
// See MarshalTo instead.
//
// Don't confuse this function with StringBytes, which must be called
// for obtaining the underlying JSON string for the v.
func (v *Value) ToString() string {
	b := v.MarshalTo(nil)
	// It is safe converting b to string without allocation, since b is no longer
	// reachable after this line.
	return b2s(b)
}

func (v *Value) ToBytes() []byte {
	b := v.MarshalTo(nil)
	// It is safe converting b to string without allocation, since b is no longer
	// reachable after this line.
	return b
}

// Type returns the type of the v.
func (v *Value) Type() Type {
	if v.t == typeRawString {
		v.s = unescapeStringBestEffort(v.s)
		v.t = TypeString
	}
	return v.t
}

// Exists returns true if the field exists for the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
func (v *Value) Exists(key string) bool {
	v = v.Get(key)
	return v != nil
}

// Get returns value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path.
//
// The returned value is valid until Parse is called on the Parser returned v.
func (v *Value) Get(key string) *Value {
	if v == nil {
		return nil
	}

	if key == "" {
		return v
	}

	keys := strings.Split(key, ".")
	for _, key := range keys {
		if v.t == TypeObject {
			v = v.o.Get(key)
			if v == nil {
				return nil
			}
		} else if v.t == TypeArray {
			n, err := strconv.Atoi(key)
			if err != nil || n < 0 || n >= len(v.a) {
				return nil
			}
			v = v.a[n]
		} else {
			return nil
		}
	}
	return v
}

// GetObject returns object value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned object is valid until Parse is called on the Parser returned v.
func (v *Value) GetObject(key string) *Object {
	v = v.Get(key)
	if v == nil || v.t != TypeObject {
		return nil
	}
	return &v.o
}

// GetArray returns array value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned array is valid until Parse is called on the Parser returned v.
func (v *Value) GetArray(key string) []*Value {
	v = v.Get(key)
	if v == nil || v.t != TypeArray {
		return nil
	}
	return v.a
}

// GetFloat64 returns float64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetFloat64(key string) float64 {
	v = v.Get(key)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseBestEffort(v.s)
}

// GetInt returns int value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetInt(key string) int {
	v = v.Get(key)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	n := fastfloat.ParseInt64BestEffort(v.s)
	nn := int(n)
	if int64(nn) != n {
		return 0
	}
	return nn
}

// GetUint returns uint value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetUint(key string) uint {
	v = v.Get(key)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	n := fastfloat.ParseUint64BestEffort(v.s)
	nn := uint(n)
	if uint64(nn) != n {
		return 0
	}
	return nn
}

// GetInt64 returns int64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetInt64(key string) int64 {
	v = v.Get(key)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseInt64BestEffort(v.s)
}

// GetUint64 returns uint64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetUint64(key string) uint64 {
	v = v.Get(key)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseUint64BestEffort(v.s)
}

// GetStringBytes returns string value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned string is valid until Parse is called on the Parser returned v.
func (v *Value) GetStringBytes(key string) []byte {
	v = v.Get(key)
	if v == nil || v.Type() != TypeString {
		return nil
	}
	return s2b(v.s)
}

func (v *Value) GetString(key string) string {
	v = v.Get(key)
	if v == nil || v.Type() != TypeString {
		return ""
	}
	return string(s2b(v.s))
}

// GetBool returns bool value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// false is returned for non-existing keys path or for invalid value type.
func (v *Value) GetBool(key string) bool {
	v = v.Get(key)
	if v != nil && v.t == TypeTrue {
		return true
	}
	return false
}

// Object returns the underlying JSON object for the v.
//
// The returned object is valid until Parse is called on the Parser returned v.
//
// Use GetObject if you don't need error handling.
func (v *Value) Object() *Object {
	if v.t != TypeObject {
		return nil
	}
	return &v.o
}

// Array returns the underlying JSON array for the v.
//
// The returned array is valid until Parse is called on the Parser returned v.
//
// Use GetArray if you don't need error handling.
func (v *Value) Array() []*Value {
	if v.t != TypeArray {
		return nil
	}
	return v.a
}

// StringBytes returns the underlying JSON string for the v.
//
// The returned string is valid until Parse is called on the Parser returned v.
//
// Use GetStringBytes if you don't need error handling.
func (v *Value) StringBytes() []byte {
	if v.Type() != TypeString {
		return nil
	}
	return s2b(v.s)
}

func (v *Value) String() string {
	if v.Type() != TypeString {
		return ""
	}
	return string(s2b(v.s))
}

// Float64 returns the underlying JSON number for the v.
//
// Use GetFloat64 if you don't need error handling.
func (v *Value) Float64() float64 {
	if v.Type() != TypeNumber {
		return 0
	}
	ret, err := fastfloat.Parse(v.s)
	if err != nil {
		return 0
	}
	return ret
}

// Int returns the underlying JSON int for the v.
//
// Use GetInt if you don't need error handling.
func (v *Value) Int() int {
	if v.Type() != TypeNumber {
		return 0
	}
	n, err := fastfloat.ParseInt64(v.s)
	if err != nil {
		return 0
	}
	nn := int(n)
	if int64(nn) != n {
		return 0
	}
	return nn
}

// Uint returns the underlying JSON uint for the v.
//
// Use GetInt if you don't need error handling.
func (v *Value) Uint() uint {
	if v.Type() != TypeNumber {
		return 0
	}
	n, err := fastfloat.ParseUint64(v.s)
	if err != nil {
		return 0
	}
	nn := uint(n)
	if uint64(nn) != n {
		return 0
	}
	return nn
}

// Int64 returns the underlying JSON int64 for the v.
//
// Use GetInt64 if you don't need error handling.
func (v *Value) Int64() int64 {
	if v.Type() != TypeNumber {
		return 0
	}
	ret, err := fastfloat.ParseInt64(v.s)
	if err != nil {
		return 0
	}
	return ret
}

// Uint64 returns the underlying JSON uint64 for the v.
//
// Use GetInt64 if you don't need error handling.
func (v *Value) Uint64() uint64 {
	if v.Type() != TypeNumber {
		return 0
	}
	ret, err := fastfloat.ParseUint64(v.s)
	if err != nil {
		return 0
	}
	return ret
}

// Bool returns the underlying JSON bool for the v.
//
// Use GetBool if you don't need error handling.
func (v *Value) Bool() bool {
	if v.t == TypeTrue {
		return true
	}
	if v.t == TypeFalse {
		return false
	}
	return false
}
