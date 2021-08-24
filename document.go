package uyaml

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
)

type Document struct {
	Value *yaml.Node
}

// DigItem attempts to retrieve an item in the provided path. Returns
// a boolean indicating if an item was found, the found item, or an error,
// if parsing the provided path fails.
func (y Document) DigItem(path string) (ok bool, val *Element, err error) {
	return dig(path, y.Value)
}

// MustDigItem works just like DigItem, but panics in case the provided path
// can't be parsed or in case an item cannot be retrieved.
func (y Document) MustDigItem(path string) *Element {
	return mustDig(path, y.Value)
}

// Remove removes the item under a given path. Returns the modified structure
// or an error, in case the path cannot be parsed.
func (y Document) Remove(path string) (obj interface{}, err error) {
	if path == "" {
		return nil, fmt.Errorf("empty path provided to Remove")
	}

	ok, v, err := y.DigItem(path)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("could not find item for path %s", path)
	}

	if err = v.Remove(); err != nil {
		return nil, err
	}

	ok, concreteValue := v.Interface()
	if !ok {
		return nil, fmt.Errorf("removed item, but could not obtain its concrete value")
	}

	return concreteValue, nil
}

// MustRemove works just like Remove, but panics in case the provided path
// can't be parsed or in case an error occurs.
func (y Document) MustRemove(path string) (obj interface{}) {
	if path == "" {
		panic("empty path provided to MustRemove")
	}
	obj, err := y.Remove(path)
	if err != nil {
		panic(err)
	}
	return obj
}

// Set sets a given value to the provided path. Structures are automatically
// created in case they don't yet exist. Returns a copy of the current structure
// containing the provided value under the specified path, or an error in case
// the path cannot be parsed.
func (y Document) Set(path string, value interface{}) (obj *Element, err error) {
	if path == "" {
		return nil, fmt.Errorf("empty path provided to Set")
	}
	return set(y.Value, path, value)
}

// MustSet works just like Set, but panics in case the provided path
// can't be parsed or in case an error occurs.
func (y Document) MustSet(path string, value interface{}) *Element {
	if path == "" {
		panic("empty path provided to MustSet")
	}
	obj, err := set(y.Value, path, value)
	if err != nil {
		panic(err)
	}
	return obj
}

// Encode encodes the underlying value into a YAML representation
func (y Document) Encode() ([]byte, error) {
	return yaml.Marshal(y.Value)
}

// EncodeIndent encodes the underlying value into a YAML representation using
// the provided indentation level.
func (y Document) EncodeIndent(indent int) ([]byte, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(indent)
	if err := enc.Encode(y.Value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
