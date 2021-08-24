package uyaml

import (
	"fmt"
	"strings"
)

type parseState int

const (
	parseStateKey parseState = iota

	// Selectors
	parseStateSelectorKey       // After leftParen found right after a dot
	parseStateSelectorOpenQuote // Quote, after parseStateSelectorKey
	parseStateSelectorValue     // Anything until unescaped quote
	parseStateSelectorEnd       // rightParen, after parseStateSelectorValue

	// Required after a parseStateSelectorEnd, just to ensure we have a dot.
	parseStateMatchDot
)

const (
	dot        = '.'
	leftParen  = '('
	rightParen = ')'
	equal      = '='
	quote      = '\''
	escape     = '\\'
)

type pathKey string
type pathSelector struct {
	Key   string
	Value string
}

func parsePath(path string) ([]interface{}, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty string provided to parsePath")
	}

	var state parseState
	var tmpString []rune
	var constructed []interface{}
	var escaping bool
	var tmpKV pathSelector
	apnd := func(r rune) {
		tmpString = append(tmpString, r)
		escaping = r == escape
	}

	for pos, c := range path {
		switch state {
		case parseStateKey:
			if c == dot && !escaping {
				constructed = append(constructed, pathKey(tmpString))
				tmpString = tmpString[:0]
				break
			} else if c == leftParen && !escaping {
				// We should have a dot before opening parens
				if len(tmpString) > 0 {
					return nil, makeError("unexpected '('", path, pos)
				}
				state = parseStateSelectorKey
				break
			}
			apnd(c)
		case parseStateSelectorKey:
			if c == equal && !escaping {
				if len(tmpString) == 0 {
					// Equal before value?
					return nil, makeError("unexpected '='", path, pos)
				}
				tmpKV.Key = string(tmpString)
				tmpString = tmpString[:0]
				state = parseStateSelectorOpenQuote
				break
			} else if (c == leftParen || c == rightParen || c == dot) && !escaping {
				return nil, makeError("unexpected token", path, pos)
			}
			apnd(c)

		case parseStateSelectorOpenQuote:
			// Just check for an opening quote.
			if c != quote {
				return nil, makeError("expected quote", path, pos)
			}
			state = parseStateSelectorValue

		case parseStateSelectorValue:
			if c == quote && !escaping {
				tmpKV.Value = string(tmpString)
				tmpString = tmpString[:0]
				state = parseStateSelectorEnd
				break
			}
			apnd(c)
		case parseStateSelectorEnd:
			if c == rightParen {
				constructed = append(constructed, tmpKV)
				tmpKV.Key = ""
				tmpKV.Value = ""
				state = parseStateMatchDot
				break
			}
			return nil, makeError("expected ')'", path, pos)
		case parseStateMatchDot:
			if c != dot {
				return nil, makeError("expected EOF or '.'", path, pos)
			}
			state = parseStateKey
		default:
			return nil, bug("unexpected parser state")
		}
	}

	switch state {
	case parseStateKey, parseStateMatchDot, parseStateSelectorKey:
		if len(tmpString) > 0 {
			constructed = append(constructed, pathKey(tmpString))
		}
	default:
		return nil, makeError("unexpected EOF", path, len(path)-1)
	}

	return constructed, nil
}

func makeError(message, path string, pos int) error {
	return fmt.Errorf("could not parse:\n%s\n%s^ %s", path, strings.Repeat(" ", pos), message)
}
