package router

import (
	"reflect"
	"testing"
)

var matcherTestData = []struct {
	name     string
	pattern  string
	path     string
	match    bool
	captures map[string]string
	error
}{
	{
		"slash",
		"/",
		"/",
		true,
		map[string]string{},
		nil,
	},
	{
		"slash mismatch",
		"/",
		"/foo",
		false,
		nil,
		nil,
	},
	{
		"basic",
		"/foo",
		"/foo",
		true,
		map[string]string{},
		nil,
	},
	{
		"basic mismatch",
		"/foo",
		"/bar",
		false,
		nil,
		nil,
	},
	{
		"basic underrun",
		"/foo",
		"/",
		false,
		nil,
		nil,
	},
	{
		"basic overrun",
		"/foo",
		"/foo/bar",
		false,
		nil,
		nil,
	},
	{
		"capture",
		"/:foo",
		"/bar",
		true,
		map[string]string{"foo": "bar"},
		nil,
	},
	{
		"user id",
		"/users/:user_id",
		"/users/abc123",
		true,
		map[string]string{"user_id": "abc123"},
		nil,
	},
	{
		"optional match",
		"(/users/:user_id)/foo",
		"/users/abc123/foo",
		true,
		map[string]string{"user_id": "abc123"},
		nil,
	},
	{
		"optional mismatch",
		"(/users/:user_id)/foo",
		"/foo",
		true,
		map[string]string{},
		nil,
	},
	{
		"nested optional",
		"(/users/(:user_id))/foo",
		"/foo",
		false,
		nil,
		errNestedOptional,
	},
	{
		"unbalanced start optional",
		"(/users/:user_id/foo",
		"/foo",
		false,
		nil,
		errUnmatchedOptionals,
	},
	{
		"unbalanced end optional",
		"(/users/:user_id))/foo",
		"/foo",
		false,
		nil,
		errUnexpectedEndOptional,
	},
}

func Test_pattern_match(t *testing.T) {
	for _, test := range matcherTestData {
		pat, err := ParsePattern(test.pattern)
		if test.error != err {
			t.Errorf("[%v] Parse error mismatch.  Expected '%v'. Got '%v'.", test.name, test.error, err)
		}
		if err != nil {
			continue
		}
		match, captures := pat.match(test.path)
		if test.match != match {
			t.Errorf("[%v] Match result mismatch.  Expected '%v'. Got '%v'.", test.name, test.match, match)
		}
		if !reflect.DeepEqual(captures, test.captures) {
			t.Errorf("[%v] Captures don't match. Expected '%v'. Got '%v'", test.name, test.captures, captures)
		}
	}

}
