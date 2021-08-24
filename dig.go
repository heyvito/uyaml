package uyaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

func dig(path string, n *yaml.Node) (bool, *Element, error) {
	if path == "" {
		return false, nil, fmt.Errorf("empty path provided to DigItem")
	}

	ok, obj, err := search(path, n)
	if !ok || err != nil {
		return ok, nil, err
	}
	return ok, obj, err
}

func mustDig(path string, n *yaml.Node) *Element {
	if path == "" {
		panic("empty path provided to MustDigItem")
	}
	ok, v, err := dig(path, n)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("item not found")
	}
	return v
}
