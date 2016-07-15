package router

import (
	"reflect"
	"testing"
)

var tokenizeTestData = []struct {
	name   string
	input  string
	tokens []token
	error
}{
	{
		"slash",
		"/",
		[]token{
			token{tokenSlash, "/"},
		},
		nil,
	},
	{
		"simple",
		"/foo",
		[]token{
			token{tokenSlash, "/"},
			token{tokenLiteral, "foo"},
		},
		nil,
	},
	{
		"capture",
		"/:foo",
		[]token{
			token{tokenSlash, "/"},
			token{tokenCapture, "foo"},
		},
		nil,
	},
}

func Test_tokenizePattern(t *testing.T) {
	for _, test := range tokenizeTestData {
		res, err := tokenizePattern(test.input)
		if !reflect.DeepEqual(res, test.tokens) {
			t.Errorf("[%v] Tokens don't match. Expected %v. Got %v", test.name, test.tokens, res)
		}
		if test.error != err {
			t.Errorf("[%v] Errors don't match. Expected %v. Got %v", test.name, test.error, err)
		}
	}
}
