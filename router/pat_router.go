package router

import (
	"fmt"
)

type tokenType int
type token struct {
	tokenType
	raw string
}

func (t token) empty() bool {
	return t.tokenType == tokenUnknown
}

const (
	tokenUnknown tokenType = iota
	tokenSlash
	tokenDot
	tokenLiteral
	tokenCapture
	tokenBeginOptional
	tokenEndOptional
)

type PatternRouter struct {
	Patterns []Pattern
}

type Pattern struct {
}

func ParsePattern(pat string) (*Pattern, error) {
	return nil, nil
}

func tokenizePattern(pat string) ([]token, error) {
	var tokens = make([]token, 0)
	var emptyToken token
	var currentToken token
	for i := 0; i < len(pat); i++ {
		var char = pat[i : i+1]
		switch char {
		case "/":
			if !currentToken.empty() {
				tokens = append(tokens, currentToken)
			}
			tokens = append(tokens, token{tokenSlash, "/"})
			currentToken = emptyToken
		case ":":
			if currentToken.empty() {
				currentToken = token{tokenCapture, ""}
			} else {
				return nil, fmt.Errorf("invalid character %v at %v", char, i)
			}
		default:
			if currentToken.empty() {
				currentToken = token{tokenLiteral, ""}
			}
			currentToken.raw += char
		}
	}
	if !currentToken.empty() {
		tokens = append(tokens, currentToken)
	}
	return tokens, nil
}
