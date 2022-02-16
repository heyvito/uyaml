package uyaml

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"strconv"
)

var yamlBoolTrue = []string{"y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON"}
var yamlBoolFalse = []string{"n", "N", "no", "No", "NO", "false", "False", "FALSE", "off", "Off", "OFF"}
var yamlBool = append(yamlBoolTrue, yamlBoolFalse...)

type Kind int

func (k Kind) Is(o Kind) bool {
	return k&o == o
}

func (k Kind) String() string {
	if v, ok := kindString[k]; ok {
		return v
	}
	return "!invalid?"
}

const (
	KindInvalid   Kind = 1 << iota
	KindString         // string
	KindFloat          // float64
	KindInt            // int
	KindBool           // bool
	KindMap            // map[string]interface{}
	KindSlice          // []interface{}
	KindInterface      // interface{}
	KindNull           // nil
)

const (
	KindSliceString = KindSlice | KindString    // []string
	KindSliceFloat  = KindSlice | KindFloat     // []float64
	KindSliceInt    = KindSlice | KindInt       // []int
	KindSliceBool   = KindSlice | KindBool      // []bool
	KindSliceMap    = KindSlice | KindMap       // []map[string]interface{}
	KindSliceMixed  = KindSlice | KindInterface // []interface{}
)

var kindString = map[Kind]string{
	KindInvalid:     "Invalid",
	KindString:      "String",
	KindFloat:       "Float",
	KindInt:         "Int",
	KindBool:        "Bool",
	KindMap:         "Map",
	KindSlice:       "Slice",
	KindInterface:   "Interface",
	KindSliceString: "SliceString",
	KindSliceFloat:  "SliceFloat",
	KindSliceInt:    "SliceInt",
	KindSliceBool:   "SliceBool",
	KindSliceMap:    "SliceMap",
	KindSliceMixed:  "SliceMixed",
	KindNull:        "Null",
}

var tagToKind = map[string]Kind{
	"!!str":   KindString,
	"!!seq":   KindSlice,
	"!!map":   KindMap,
	"!!bool":  KindBool,
	"!!int":   KindInt,
	"!!float": KindFloat,
	"!!null":  KindNull,
}

func element(n *yaml.Node) *Element {
	return &Element{
		value:  n,
		parent: nil,
	}
}

type Element struct {
	value  *yaml.Node
	parent *Element
}

// Remove removes the receiver from its parent. Returns an error in case the
// item cannot be removed
func (e *Element) Remove() error {
	p := e.parent
	if p.Kind().Is(KindMap) || p.Kind().Is(KindSlice) {
		idx, err := e.indexInParent()
		if err != nil {
			return err
		}
		e.parent.value.Content = append(e.parent.value.Content[0:idx], e.parent.value.Content[idx+1:]...)
		return nil
	}

	return fmt.Errorf("cannot remove element from parent of kind %s", p.Kind())
}

func (e *Element) indexInParent() (int, error) {
	p := e.parent
	itemIdx := -1
	for i, v := range p.value.Content {
		if v == e.value {
			itemIdx = i
			break
		}
	}
	if itemIdx == -1 {
		return -1, fmt.Errorf("couldn't find element %p in parent", e.value)
	}
	return itemIdx, nil
}

// Replace replaces the receiver in its parent, returning the new Element
// placed on its previous value.
func (e *Element) Replace(newValue interface{}) (*Element, error) {
	idx, err := e.indexInParent()
	if err != nil {
		return nil, err
	}
	n, err := buildNode(newValue)
	if err != nil {
		return nil, err
	}
	e.parent.value.Content[idx] = n
	return element(n), nil
}

// Decode decodes the receiver into a provided struct pointer
func (e *Element) Decode(into interface{}) error {
	return e.value.Decode(into)
}

// Kind returns the receiver's element Kind
func (e *Element) Kind() Kind {
	k, ok := tagToKind[e.value.Tag]
	if !ok {
		return KindInvalid
	}

	// Seq?
	if k == KindSlice {
		nodeKind := ""
		for _, v := range e.value.Content {
			if nodeKind == "" {
				nodeKind = v.Tag
				continue
			}
			if v.Tag != nodeKind {
				return KindSliceMixed
			}
		}
		if k, ok := tagToKind[nodeKind]; ok {
			return KindSlice | k
		}
		return KindSliceMixed
	}

	return k
}

// Dig attempts to retrieve an item in the provided path. Returns
// a boolean indicating if an item was found, the found item, or an error,
// if parsing the provided path fails.
func (e *Element) Dig(path string) (bool, *Element, error) {
	return dig(path, e.value)
}

// MustDig works just like Dig, but panics in case the provided path
// can't be parsed or in case an item cannot be retrieved.
func (e *Element) MustDig(path string) *Element {
	return mustDig(path, e.value)
}

// String returns a boolean indicating whether the receiver can be coerced into
// a string value, and if positive, the receiver's value
func (e *Element) String() (bool, string) {
	if e.Kind() != KindString {
		return false, ""
	}
	return true, e.value.Value
}

func (e *Element) anyOf(kinds ...Kind) bool {
	kind := e.Kind()
	for _, k := range kinds {
		if k == kind {
			return true
		}
	}
	return false
}

// Float returns a boolean indicating whether the receiver can be coerced into
// a float64 value, and if positive, the receiver's value
func (e *Element) Float() (bool, float64) {
	switch e.Kind() {
	case KindInt:
		ok, v := e.Int()
		return ok, float64(v)
	case KindFloat:
		v, _ := strconv.ParseFloat(e.value.Value, 64)
		return true, v
	}
	return false, 0
}

// Int returns a boolean indicating whether the receiver can be coerced into
// an int64 value, and if positive, the receiver's value
func (e *Element) Int() (bool, int64) {
	switch e.Kind() {
	case KindInt:
		v, _ := strconv.ParseInt(e.value.Value, 10, 64)
		return true, v
	case KindFloat:
		ok, v := e.Float()
		return ok, int64(v)
	}
	return false, 0
}

// Bool returns a boolean indicating whether the receiver can be coerced into
// a boolean value, and if positive, the receiver's value
func (e *Element) Bool() (bool, bool) {
	if e.Kind() != KindBool {
		return false, false
	}

	ok := false
	for _, v := range yamlBool {
		if v == e.value.Value {
			ok = true
		}
	}

	if !ok {
		return true, false
	}

	for _, v := range yamlBoolTrue {
		if v == e.value.Value {
			return true, true
		}
	}

	return true, false
}

// Map returns a boolean indicating whether the receiver can be coerced into
// a map[string]interface{}, and if positive, the receiver's value
func (e *Element) Map() (bool, map[string]interface{}) {
	if e.Kind() != KindMap {
		return false, nil
	}

	m := map[string]interface{}{}
	var k string
	var ok bool
	for i, v := range e.value.Content {
		if i%2 == 0 {
			k = v.Value
		} else {
			ok, m[k] = element(v).Interface()
			if !ok {
				return false, nil
			}
		}
	}

	return true, m
}

// Interface returns a boolean indicating whether the receiver can be coerced
// into a generic interface{} value, and if positive, the receiver's value
func (e *Element) Interface() (bool, interface{}) {
	switch e.Kind() {
	case KindString:
		return e.String()
	case KindFloat:
		return e.Float()
	case KindInt:
		return e.Int()
	case KindBool:
		return e.Bool()
	case KindMap:
		return e.Map()
	case KindInterface:
	// ?
	case KindNull:
		return true, nil
	}

	if e.Kind()&KindSlice == KindSlice {
		return e.InterfaceSlice()
	}
	return false, nil
}

// StringSlice returns a boolean indicating whether the receiver can be coerced
// into a []string value, and if positive, the receiver's value
func (e *Element) StringSlice() (bool, []string) {
	if e.Kind() != KindSliceString {
		return false, nil
	}

	arr := make([]string, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		arr = append(arr, v.Value)
	}

	return true, arr
}

// FloatSlice returns a boolean indicating whether the receiver can be coerced
// into a []float value, and if positive, the receiver's value
func (e *Element) FloatSlice() (bool, []float64) {
	if e.Kind() != KindSliceFloat {
		return false, nil
	}

	arr := make([]float64, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		arr = append(arr, element(v).MustFloat())
	}

	return true, arr
}

// IntSlice returns a boolean indicating whether the receiver can be coerced
// into a []int64 value, and if positive, the receiver's value
func (e *Element) IntSlice() (bool, []int64) {
	if e.Kind() != KindSliceInt {
		return false, nil
	}

	arr := make([]int64, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		arr = append(arr, element(v).MustInt())
	}

	return true, arr
}

// BoolSlice returns a boolean indicating whether the receiver can be coerced
// into a []bool value, and if positive, the receiver's value
func (e *Element) BoolSlice() (bool, []bool) {
	if e.Kind() != KindSliceBool {
		return false, nil
	}

	arr := make([]bool, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		arr = append(arr, element(v).MustBool())
	}

	return true, arr
}

// MapSlice returns a boolean indicating whether the receiver can be coerced
// into a []map[string]interface{} value, and if positive, the receiver's value
func (e *Element) MapSlice() (bool, []map[string]interface{}) {
	if e.Kind() != KindSliceMap {
		return false, nil
	}

	arr := make([]map[string]interface{}, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		ok, f := element(v).Map()
		if !ok {
			return false, nil
		}
		arr = append(arr, f)
	}

	return true, arr
}

// InterfaceSlice returns a boolean indicating whether the receiver can be
// coerced into a []interface{} value, and if positive, the receiver's value
func (e *Element) InterfaceSlice() (bool, []interface{}) {
	if e.Kind()&KindSlice != KindSlice {
		return false, nil
	}

	arr := make([]interface{}, 0, len(e.value.Content))
	for _, v := range e.value.Content {
		ok, val := element(v).Interface()
		if !ok {
			return false, nil
		}
		arr = append(arr, val)
	}

	return true, arr
}

func (e *Element) IsNull() bool {
	return e.Kind() == KindNull
}

// Encode encodes the underlying value into a YAML representation
func (e *Element) Encode() ([]byte, error) {
	return yaml.Marshal(findRoot(e))
}

// EncodeIndent encodes the underlying value into a YAML representation using
// the provided indentation level.
func (e *Element) EncodeIndent(indent int) ([]byte, error) {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(indent)
	if err := enc.Encode(findRoot(e)); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
