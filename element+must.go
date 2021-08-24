package uyaml

func (e *Element) MustString() string {
	ok, v := e.String()
	if !ok {
		panic("MustString: could not convert value to string")
	}
	return v
}

func (e *Element) MustFloat() float64 {
	ok, v := e.Float()
	if !ok {
		panic("MustFloat: could not convert value to float64")
	}
	return v
}

func (e *Element) MustInt() int64 {
	ok, v := e.Int()
	if !ok {
		panic("MustInt: could not convert value to int64")
	}
	return v
}

func (e *Element) MustBool() bool {
	ok, v := e.Bool()
	if !ok {
		panic("MustBool: could not convert value to bool")
	}
	return v
}

func (e *Element) MustMap() map[string]interface{} {
	ok, v := e.Map()
	if !ok {
		panic("MustMap: could not convert value to map[string]interface{}")
	}
	return v
}

func (e *Element) MustSlice() []interface{} {
	ok, v := e.InterfaceSlice()
	if !ok {
		panic("MustSlice: could not convert value to []interface{}")
	}
	return v
}
