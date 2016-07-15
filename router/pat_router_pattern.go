package router

import (
	"errors"
)

var (
	errNestedOptional        = errors.New("nested optionals are not supported")
	errUnexpectedEndOptional = errors.New("unexpected end optional")
	errUnmatchedOptionals    = errors.New("unmatched optionals")
)

type Pattern struct {
	raw    string
	tokens []token
}

func ParsePattern(pat string) (*Pattern, error) {
	if t, e := tokenizePattern(pat); e == nil {
		p := &Pattern{pat, t}
		if e = p.validate(); e == nil {
			return p, nil
		} else {
			return nil, e
		}
	} else {
		return nil, e
	}
}

func (p *Pattern) validate() error {
	var optionCount int = 0

	for _, token := range p.tokens {
		switch token.tokenType {
		case tokenBeginOptional:
			optionCount++
		case tokenEndOptional:
			optionCount--
		}

		if optionCount > 1 {
			return errNestedOptional
		}
		if optionCount < 0 {
			return errUnexpectedEndOptional
		}
	}
	if optionCount != 0 {
		return errUnmatchedOptionals
	}

	return nil
}

func (p *Pattern) match(path string) (bool, map[string]string) {
	matches := make(map[string]string)

	var optional bool
	var optionalStartPath string

	for i := 0; i < len(p.tokens); i++ {
		token := p.tokens[i]
		if match, length := token.match(path); match {
			switch token.tokenType {
			case tokenBeginOptional:
				optional = true
				optionalStartPath = path
			case tokenEndOptional:
				optional = false
			case tokenCapture:
				matches[token.raw] = path[:length]
			}
			path = path[length:]
		} else {
			if optional {
				// optional match failed.
				// ffwd to end of optional section and restart
				path = optionalStartPath
				optional = false
				for p.tokens[i].tokenType != tokenEndOptional {
					i++
				}
			} else {
				return false, nil
			}
		}
	}
	if len(path) > 0 {
		return false, nil
	}

	return true, matches
}
