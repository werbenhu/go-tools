package kson

import (
	ji "github.com/json-iterator/go"
)

var js = ji.ConfigCompatibleWithStandardLibrary

type RawMessage []byte

func (b *RawMessage) MarshalJSON() ([]byte, error) {
	return []byte(*b), nil
}

func (b *RawMessage) UnmarshalJSON(input []byte) error {
	*b = RawMessage(input)
	return nil
}

func Marshal(object interface{}) ([]byte, error) {
	return js.Marshal(object)
}

func Unmarshal(data []byte, v interface{}) error {
	return js.Unmarshal(data, v)
}
