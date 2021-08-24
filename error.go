package uyaml

import "fmt"

type ErrBug struct {
	msg string
}

func (e ErrBug) Error() string {
	return e.msg
}

func bug(format string, a ...interface{}) error {
	return ErrBug{msg: "BUG: " + fmt.Sprintf(format, a...)}
}
