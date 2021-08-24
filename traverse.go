package uyaml

import "gopkg.in/yaml.v3"

func search(path string, obj *yaml.Node) (bool, *Element, error) {
	composed, err := parsePath(path)
	if err != nil {
		return false, nil, err
	}

	ok, res := applySearch(composed, obj)
	return ok, res, nil
}

func applyPathKey(t pathKey, obj *yaml.Node) (*yaml.Node, bool) {
	if obj.Kind == yaml.DocumentNode {
		for _, v := range obj.Content {
			if n, ok := applyPathKey(t, v); ok {
				return n, ok
			}
		}
	} else if obj.Kind == yaml.MappingNode {
		takeNext := false
		for i, v := range obj.Content {
			if takeNext {
				return v, true
			}
			if i % 2 != 0 {
				continue
			}
			if v.Value == string(t) {
				takeNext = true
			}
		}
	}
	return nil, false
}

func applyPathSelector(sel pathSelector, obj *yaml.Node) (*yaml.Node, bool) {
	if obj.Kind == yaml.DocumentNode {
		for _, v := range obj.Content {
			if n, ok := applyPathSelector(sel, v); ok {
				return n, ok
			}
		}
	} else if obj.Kind == yaml.SequenceNode {
		for _, v := range obj.Content {
			if selRes, ok := applyPathKey(pathKey(sel.Key), v); ok {
				if selRes.Value == sel.Value {
					return v, ok
				}
			}
		}
	}

	return nil, false
}

func applySearch(path []interface{}, obj *yaml.Node) (bool, *Element) {
	el := element(obj)
	for _, v := range path {
		switch t := v.(type) {
		case pathKey:
			if v, ok := applyPathKey(t, obj); ok {
				nel := element(v)
				nel.parent = el
				obj = v
				el = nel
			} else {
				return false, nil
			}
		case pathSelector:
			if v, ok := applyPathSelector(t, obj); ok {
				nel := element(v)
				nel.parent = el
				obj = v
				el = nel
			} else {
				return false, nil
			}
		}
	}
	return true, el
}
