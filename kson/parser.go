package kson

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Parser parses JSON.
//
// Parser may be re-used for subsequent parsing.
//
// Parser cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Parser struct {
	// b contains working copy of the string to be parsed.
	b []byte

	// c is a cache for json values.
	c cache
}

// Parse parses s containing JSON.
//
// The returned value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed.
func (p *Parser) Parse(s string) (*Value, error) {
	s = skipWS(s)
	p.b = append(p.b[:0], s...)
	p.c.reset()

	v, tail, err := parseValue(b2s(p.b), &p.c, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, startEndString(tail))
	}
	tail = skipWS(tail)
	if len(tail) > 0 {
		return nil, fmt.Errorf("unexpected tail: %q", startEndString(tail))
	}
	return v, nil
}

// ParseBytes parses b containing JSON.
//
// The returned Value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed.
func (p *Parser) ParseBytes(b []byte) (*Value, error) {
	return p.Parse(b2s(b))
}

type cache struct {
	vs []Value
}

func (c *cache) reset() {
	c.vs = c.vs[:0]
}

func (c *cache) getValue() *Value {
	if cap(c.vs) > len(c.vs) {
		c.vs = c.vs[:len(c.vs)+1]
	} else {
		c.vs = append(c.vs, Value{})
	}
	// Do not reset the value, since the caller must properly init it.
	return &c.vs[len(c.vs)-1]
}

func skipWS(s string) string {
	if len(s) == 0 || s[0] > 0x20 {
		// Fast path.
		return s
	}
	return skipWSSlow(s)
}

func skipWSSlow(s string) string {
	if len(s) == 0 || s[0] != 0x20 && s[0] != 0x0A && s[0] != 0x09 && s[0] != 0x0D {
		return s
	}
	for i := 1; i < len(s); i++ {
		if s[i] != 0x20 && s[i] != 0x0A && s[i] != 0x09 && s[i] != 0x0D {
			return s[i:]
		}
	}
	return ""
}

type kv struct {
	k string
	v *Value
}

// MaxDepth is the maximum depth for nested JSON.
const MaxDepth = 300

func parseValue(s string, c *cache, depth int) (*Value, string, error) {
	if len(s) == 0 {
		return nil, s, fmt.Errorf("cannot parse empty string")
	}
	depth++
	if depth > MaxDepth {
		return nil, s, fmt.Errorf("too big depth for the nested JSON; it exceeds %d", MaxDepth)
	}

	if s[0] == '{' {
		v, tail, err := parseObject(s[1:], c, depth)
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parse object: %s", err)
		}
		return v, tail, nil
	}
	if s[0] == '[' {
		v, tail, err := parseArray(s[1:], c, depth)
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parse array: %s", err)
		}
		return v, tail, nil
	}
	if s[0] == '"' {
		ss, tail, err := parseRawString(s[1:])
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parse string: %s", err)
		}
		v := c.getValue()
		v.t = typeRawString
		v.s = ss
		return v, tail, nil
	}
	if s[0] == 't' {
		if len(s) < len("true") || s[:len("true")] != "true" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueTrue, s[len("true"):], nil
	}
	if s[0] == 'f' {
		if len(s) < len("false") || s[:len("false")] != "false" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueFalse, s[len("false"):], nil
	}
	if s[0] == 'n' {
		if len(s) < len("null") || s[:len("null")] != "null" {
			// Try parsing NaN
			if len(s) >= 3 && strings.EqualFold(s[:3], "nan") {
				v := c.getValue()
				v.t = TypeNumber
				v.s = s[:3]
				return v, s[3:], nil
			}
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueNull, s[len("null"):], nil
	}

	ns, tail, err := parseRawNumber(s)
	if err != nil {
		return nil, tail, fmt.Errorf("cannot parse number: %s", err)
	}
	v := c.getValue()
	v.t = TypeNumber
	v.s = ns
	return v, tail, nil
}

func parseArray(s string, c *cache, depth int) (*Value, string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing ']'")
	}

	if s[0] == ']' {
		v := c.getValue()
		v.t = TypeArray
		v.a = v.a[:0]
		return v, s[1:], nil
	}

	a := c.getValue()
	a.t = TypeArray
	a.a = a.a[:0]
	for {
		var v *Value
		var err error

		s = skipWS(s)
		v, s, err = parseValue(s, c, depth)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse array value: %s", err)
		}
		a.a = append(a.a, v)

		s = skipWS(s)
		if len(s) == 0 {
			return nil, s, fmt.Errorf("unexpected end of array")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == ']' {
			s = s[1:]
			return a, s, nil
		}
		return nil, s, fmt.Errorf("missing ',' after array value")
	}
}

func parseObject(s string, c *cache, depth int) (*Value, string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing '}'")
	}

	if s[0] == '}' {
		v := c.getValue()
		v.t = TypeObject
		v.o.reset()
		return v, s[1:], nil
	}

	o := c.getValue()
	o.t = TypeObject
	o.o.reset()
	for {
		var err error
		kv := o.o.getKV()

		// Parse key.
		s = skipWS(s)
		if len(s) == 0 || s[0] != '"' {
			return nil, s, fmt.Errorf(`cannot find opening '"" for object key`)
		}
		kv.k, s, err = parseRawKey(s[1:])
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse object key: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 || s[0] != ':' {
			return nil, s, fmt.Errorf("missing ':' after object key")
		}
		s = s[1:]

		// Parse value
		s = skipWS(s)
		kv.v, s, err = parseValue(s, c, depth)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse object value: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 {
			return nil, s, fmt.Errorf("unexpected end of object")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == '}' {
			return o, s[1:], nil
		}
		return nil, s, fmt.Errorf("missing ',' after object value")
	}
}

func escapeString(dst []byte, s string) []byte {
	if !hasSpecialChars(s) {
		// Fast path - nothing to escape.
		dst = append(dst, '"')
		dst = append(dst, s...)
		dst = append(dst, '"')
		return dst
	}

	// Slow path.
	return strconv.AppendQuote(dst, s)
}

func hasSpecialChars(s string) bool {
	if strings.IndexByte(s, '"') >= 0 || strings.IndexByte(s, '\\') >= 0 {
		return true
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
}

func unescapeStringBestEffort(s string) string {
	n := strings.IndexByte(s, '\\')
	if n < 0 {
		// Fast path - nothing to unescape.
		return s
	}

	// Slow path - unescape string.
	b := s2b(s) // It is safe to do, since s points to a byte slice in Parser.b.
	b = b[:n]
	s = s[n+1:]
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '"':
			b = append(b, '"')
		case '\\':
			b = append(b, '\\')
		case '/':
			b = append(b, '/')
		case 'b':
			b = append(b, '\b')
		case 'f':
			b = append(b, '\f')
		case 'n':
			b = append(b, '\n')
		case 'r':
			b = append(b, '\r')
		case 't':
			b = append(b, '\t')
		case 'u':
			if len(s) < 4 {
				// Too short escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				// Invalid escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			s = s[4:]
			if !utf16.IsSurrogate(rune(x)) {
				b = append(b, string(rune(x))...)
				break
			}

			// Surrogate.
			// See https://en.wikipedia.org/wiki/Universal_Character_Set_characters#Surrogates
			if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			x1, err := strconv.ParseUint(s[2:6], 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			r := utf16.DecodeRune(rune(x), rune(x1))
			b = append(b, string(r)...)
			s = s[6:]
		default:
			// Unknown escape sequence. Just store it unchanged.
			b = append(b, '\\', ch)
		}
		n = strings.IndexByte(s, '\\')
		if n < 0 {
			b = append(b, s...)
			break
		}
		b = append(b, s[:n]...)
		s = s[n+1:]
	}
	return b2s(b)
}

// parseRawKey is similar to parseRawString, but is optimized
// for small-sized keys without escape sequences.
func parseRawKey(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			// Fast path.
			return s[:i], s[i+1:], nil
		}
		if s[i] == '\\' {
			// Slow path.
			return parseRawString(s)
		}
	}
	return s, "", fmt.Errorf(`missing closing '"'`)
}

func parseRawString(s string) (string, string, error) {
	n := strings.IndexByte(s, '"')
	if n < 0 {
		return s, "", fmt.Errorf(`missing closing '"'`)
	}
	if n == 0 || s[n-1] != '\\' {
		// Fast path. No escaped ".
		return s[:n], s[n+1:], nil
	}

	// Slow path - possible escaped " found.
	ss := s
	for {
		i := n - 1
		for i > 0 && s[i-1] == '\\' {
			i--
		}
		if uint(n-i)%2 == 0 {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}
		s = s[n+1:]

		n = strings.IndexByte(s, '"')
		if n < 0 {
			return ss, "", fmt.Errorf(`missing closing '"'`)
		}
		if n == 0 || s[n-1] != '\\' {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}
	}
}

func parseRawNumber(s string) (string, string, error) {
	// The caller must ensure len(s) > 0

	// Find the end of the number.
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == 'e' || ch == 'E' || ch == '+' {
			continue
		}
		if i == 0 || i == 1 && (s[0] == '-' || s[0] == '+') {
			if len(s[i:]) >= 3 {
				xs := s[i : i+3]
				if strings.EqualFold(xs, "inf") || strings.EqualFold(xs, "nan") {
					return s[:i+3], s[i+3:], nil
				}
			}
			return "", s, fmt.Errorf("unexpected char: %q", s[:1])
		}
		ns := s[:i]
		s = s[i:]
		return ns, s, nil
	}
	return s, "", nil
}

// Object represents JSON object.
//
// Object cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Object struct {
	kvs           []kv
	keysUnescaped bool
}

func (o *Object) reset() {
	o.kvs = o.kvs[:0]
	o.keysUnescaped = false
}

// MarshalTo appends marshaled o to dst and returns the result.
func (o *Object) MarshalTo(dst []byte) []byte {
	dst = append(dst, '{')
	for i, kv := range o.kvs {
		if o.keysUnescaped {
			dst = escapeString(dst, kv.k)
		} else {
			dst = append(dst, '"')
			dst = append(dst, kv.k...)
			dst = append(dst, '"')
		}
		dst = append(dst, ':')
		dst = kv.v.MarshalTo(dst)
		if i != len(o.kvs)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, '}')
	return dst
}

// String returns string representation for the o.
//
// This function is for debugging purposes only. It isn't optimized for speed.
// See MarshalTo instead.
func (o *Object) String() string {
	b := o.MarshalTo(nil)
	// It is safe converting b to string without allocation, since b is no longer
	// reachable after this line.
	return b2s(b)
}

func (o *Object) getKV() *kv {
	if cap(o.kvs) > len(o.kvs) {
		o.kvs = o.kvs[:len(o.kvs)+1]
	} else {
		o.kvs = append(o.kvs, kv{})
	}
	return &o.kvs[len(o.kvs)-1]
}

func (o *Object) unescapeKeys() {
	if o.keysUnescaped {
		return
	}
	kvs := o.kvs
	for i := range kvs {
		kv := &kvs[i]
		kv.k = unescapeStringBestEffort(kv.k)
	}
	o.keysUnescaped = true
}

// Len returns the number of items in the o.
func (o *Object) Len() int {
	return len(o.kvs)
}

// Get returns the value for the given key in the o.
//
// Returns nil if the value for the given key isn't found.
//
// The returned value is valid until Parse is called on the Parser returned o.
func (o *Object) Get(key string) *Value {
	if !o.keysUnescaped && strings.IndexByte(key, '\\') < 0 {
		// Fast path - try searching for the key without object keys unescaping.
		for _, kv := range o.kvs {
			if kv.k == key {
				return kv.v
			}
		}
	}

	// Slow path - unescape object keys.
	o.unescapeKeys()

	for _, kv := range o.kvs {
		if kv.k == key {
			return kv.v
		}
	}
	return nil
}

// Visit calls f for each item in the o in the original order
// of the parsed JSON.
//
// f cannot hold key and/or v after returning.
func (o *Object) Visit(f func(key []byte, v *Value)) {
	if o == nil {
		return
	}

	o.unescapeKeys()

	for _, kv := range o.kvs {
		f(s2b(kv.k), kv.v)
	}
}

// Type represents JSON type.
type Type int

const (
	// TypeNull is JSON null.
	TypeNull Type = 0

	// TypeObject is JSON object type.
	TypeObject Type = 1

	// TypeArray is JSON array type.
	TypeArray Type = 2

	// TypeString is JSON string type.
	TypeString Type = 3

	// TypeNumber is JSON number type.
	TypeNumber Type = 4

	// TypeTrue is JSON true.
	TypeTrue Type = 5

	// TypeFalse is JSON false.
	TypeFalse Type = 6

	typeRawString Type = 7
)

// String returns string representation of t.
func (t Type) String() string {
	switch t {
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeTrue:
		return "true"
	case TypeFalse:
		return "false"
	case TypeNull:
		return "null"

	// typeRawString is skipped intentionally,
	// since it shouldn't be visible to user.
	default:
		panic(fmt.Errorf("BUG: unknown Value type: %d", t))
	}
}

var (
	valueTrue  = &Value{t: TypeTrue}
	valueFalse = &Value{t: TypeFalse}
	valueNull  = &Value{t: TypeNull}
)
