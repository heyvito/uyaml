package uyaml

import (
	"gopkg.in/yaml.v3"
)



func Decode(data []byte) (*Document, error) {
	var n yaml.Node
	err := yaml.Unmarshal(data, &n)
	if err != nil {
		return nil, err
	}
	return &Document{&n}, nil
}
