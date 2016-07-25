package router

import (
	"fmt"
	"strings"
)

type tokenType int
type token struct {
	tokenType
	raw string
}

func (t token) empty() bool {
	return t.tokenType == tokenUnknown
}

func (t token) match(input string) (bool, int) {
	switch t.tokenType {
	case tokenBeginOptional:
		return true, 0
	case tokenEndOptional:
		return true, 0
	case tokenCapture:
		for i := 0; i < len(input); i++ {
			if input[i:i+1] == "/" {
				return true, i
			}
		}
		return true, len(input)
	default:
		if l := len(t.raw); len(input) >= l && input[:l] == t.raw {
			return true, l
		} else {
			return false, 0
		}
	}
}

const (
	tokenUnknown tokenType = iota
	tokenSlash
	tokenDot
	tokenLiteral
	tokenCapture
	tokenWildcard
	tokenBeginOptional
	tokenEndOptional
)

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
		case "(":
			if !currentToken.empty() {
				return nil, fmt.Errorf("unexpected optional start at %v", i)
			}
			tokens = append(tokens, token{tokenBeginOptional, "("})
			currentToken = emptyToken
		case ")":
			if !currentToken.empty() {
				tokens = append(tokens, currentToken)
			}
			tokens = append(tokens, token{tokenEndOptional, ")"})
			currentToken = emptyToken
		case ".":
			// only tokenize . at the end of a url
			if strings.Contains(pat[i+1:], "/") || strings.Contains(pat[i+1:], ".") {
				if currentToken.empty() {
					currentToken = token{tokenLiteral, ""}
				}
				currentToken.raw += char
			} else {
				if !currentToken.empty() {
					tokens = append(tokens, currentToken)
				}
				tokens = append(tokens, token{tokenDot, "."})
				currentToken = emptyToken
			}
		case "*":
			if !currentToken.empty() {
				tokens = append(tokens, currentToken)
			}
			tokens = append(tokens, token{tokenWildcard, "*"})
			currentToken = emptyToken
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
