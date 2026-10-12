package envai

import "encoding/json"

type Optional[T any] struct {
	Value T
	Set   bool
	Null  bool
}

func Present[T any](value T) Optional[T] { return Optional[T]{Value: value, Set: true} }
func Null[T any]() Optional[T]           { return Optional[T]{Set: true, Null: true} }
func (v Optional[T]) IsZero() bool       { return !v.Set }
func (v Optional[T]) MarshalJSON() ([]byte, error) {
	if v.Null || !v.Set {
		return []byte("null"), nil
	}
	return json.Marshal(v.Value)
}
func (v *Optional[T]) UnmarshalJSON(b []byte) error {
	v.Set = true
	v.Null = string(b) == "null"
	if v.Null {
		var zero T
		v.Value = zero
		return nil
	}
	return json.Unmarshal(b, &v.Value)
}
