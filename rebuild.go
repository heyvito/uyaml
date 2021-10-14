package uyaml

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"reflect"
	"strconv"
)

func findRoot(el *Element) *yaml.Node {
	p := el
	for p.parent != nil {
		p = p.parent
	}
	return p.value
}

func buildNode(val interface{}) (*yaml.Node, error) {
	n := &yaml.Node{}
	switch v := val.(type) {
	case string:
		n.Tag = "!!str"
		n.Kind = yaml.ScalarNode
		n.Value = v
	case int:
		n.Tag = "!!int"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.Itoa(v)
	case int8:
		n.Tag = "!!int"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.Itoa(int(v))
	case int16:
		n.Tag = "!!int"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.Itoa(int(v))
	case int32:
		n.Tag = "!!int"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.Itoa(int(v))
	case int64:
		n.Tag = "!!int"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.Itoa(int(v))
	case float32:
		n.Tag = "!!float"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.FormatFloat(float64(v), 'f', 18, 32)
	case float64:
		n.Tag = "!!float"
		n.Kind = yaml.ScalarNode
		n.Value = strconv.FormatFloat(v, 'f', 18, 64)
	case bool:
		n.Tag = "!!bool"
		n.Kind = yaml.ScalarNode
		if v {
			n.Value = "true"
		} else {
			n.Value = "false"
		}
	default:
		t := reflect.TypeOf(val)
		reflectedValue := reflect.ValueOf(val)
		if t.Kind() == reflect.Array {
			var nodeArr []*yaml.Node
			l := reflectedValue.Len()
			for i := 0; i < l; i++ {
				n, err := buildNode(reflectedValue.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				nodeArr = append(nodeArr, n)
			}
			n.Kind = yaml.SequenceNode
			n.Content = nodeArr
		} else if t.Kind() == reflect.Slice {
			var nodeArr []*yaml.Node
			l := reflectedValue.Len()
			for i := 0; i < l; i++ {
				n, err := buildNode(reflectedValue.Index(i).Interface())
				if err != nil {
					return nil, err
				}
				nodeArr = append(nodeArr, n)
			}
			n.Kind = yaml.SequenceNode
			n.Tag = "!!seq"
			n.Content = nodeArr
		} else if t.Kind() == reflect.Map {
			keys := reflectedValue.MapKeys()
			var nodeArr []*yaml.Node
			for _, k := range keys {
				if k.Kind() != reflect.String {
					return nil, fmt.Errorf("could not create map with unsupported key type %s", k.Kind())
				}
				key, err := buildNode(k.String())
				if err != nil {
					return nil, err
				}
				nodeArr = append(nodeArr, key)
				val, err := buildNode(reflectedValue.MapIndex(k).Interface())
				if err != nil {
					return nil, err
				}
				nodeArr = append(nodeArr, val)
			}
			n.Kind = yaml.MappingNode
			n.Tag = "!!map"
			n.Content = nodeArr
		} else {
			return nil, fmt.Errorf("could not build node with type %T", val)
		}
	}

	return n, nil
}

func set(root *yaml.Node, path string, value interface{}) (*Element, error) {
	// Node exists?
	ok, v, err := search(path, root)
	if err != nil {
		return nil, err
	}
	if ok {
		if n, err := v.Replace(value); err != nil {
			return nil, err
		} else {
			return n, nil
		}
	}

	// At this point, node does not exist at some point. Iterate until we have
	// all structures in place.
	composed, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	obj := root
	el := element(root)

	for i, v := range composed {
		switch t := v.(type) {
		case pathKey:
			if v, ok := applyPathKey(t, obj); ok {
				nel := element(v)
				nel.parent = el
				obj = v
				el = nel
			} else {
				// At this point, el does not have path components for
				// composed[i:]
				if err = buildAndSet(el, composed[i:], value); err != nil {
					return nil, err
				}
				return el, nil
			}
		case pathSelector:
			if v, ok := applyPathSelector(t, obj); ok {
				nel := element(v)
				nel.parent = el
				obj = v
				el = nel
			} else {
				// At this point, el does not have path components for
				// composed[i:]
				if err = buildAndSet(el, composed[i:], value); err != nil {
					return nil, err
				}
				return el, nil
			}
		}
	}

	return nil, bug("Unexpected state for set function, should have returned from the previous loop")
}

func buildAndSet(el *Element, path []interface{}, value interface{}) error {
	e, err := buildStructure(path, value)
	if err != nil {
		return err
	}
	if el.value.Kind == yaml.DocumentNode {
		innerEl := el.value.Content[0]
		innerEl.Content = append(innerEl.Content, e.value.Content...)
	} else {
		if el.Kind() == e.Kind() {
			// ...merge?
			el.value.Content = append(el.value.Content, e.value.Content...)
		} else {
			// ...append?
			el.value.Content = append(el.value.Content, e.value)
		}
	}

	return nil
}

func buildStructure(components []interface{}, value interface{}) (*Element, error) {
	if len(components) == 0 {
		return nil, bug("buildStructure received an empty components slice")
	}

	// Take the last component, and create it.
	lastComp := components[len(components)-1]
	elementBase, err := buildNode(value)
	if err != nil {
		return nil, err
	}

	var baseValue []*yaml.Node
	if sel, ok := lastComp.(pathSelector); ok {
		k := sel.Key
		v := sel.Value
		elementBase.Content = append([]*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: k,
			},
			{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: v,
			},
		}, elementBase.Content...)
		baseValue = []*yaml.Node{
			{
				Kind:    yaml.SequenceNode,
				Tag:     "!!seq",
				Content: []*yaml.Node{elementBase},
				Column:  0,
			},
		}
	} else {
		baseValue = []*yaml.Node{
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{
						Kind:  yaml.ScalarNode,
						Tag:   "!!str",
						Value: string(lastComp.(pathKey)),
					},
					elementBase,
				},
			},
		}
	}

	components = components[:len(components)-1]
	for _, rv := range components {
		switch v := rv.(type) {
		case pathKey:
			baseValue = []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
					Content: append([]*yaml.Node{
						{
							Kind:  yaml.ScalarNode,
							Tag:   "!!str",
							Value: string(v),
						},
					},
						baseValue...),
				},
			}
		case pathSelector:
			baseValue = []*yaml.Node{
				{
					Kind:    yaml.SequenceNode,
					Tag:     "!!seq",
					Content: baseValue,
				},
			}
		}
	}

	return element(baseValue[0]), nil
}
