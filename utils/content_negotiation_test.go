package utils

import (
	"testing"
)

var contentNegotiationTestData = []struct {
	name      string
	accept    []string
	available []string
	expect    string
}{
	{
		"basic",
		[]string{"image/jpeg; q=1.0", "image/png; q=0.8"},
		[]string{"image/png", "image/jpeg"},
		"image/jpeg",
	},
	{
		"reverse",
		[]string{"image/jpeg; q=0.8", "image/png; q=1.0"},
		[]string{"image/png", "image/jpeg"},
		"image/png",
	},
	{
		"partial wild",
		[]string{"image/*; q=0.8", "*/*; q=0.1"},
		[]string{"image/png", "image/jpeg"},
		"image/png",
	},
	{
		"full wild",
		[]string{"text/*; q=0.8", "*/*; q=0.1"},
		[]string{"image/png", "image/jpeg"},
		"image/png",
	},
	{
		"no match",
		[]string{"image/jpeg; q=0.8", "image/png; q=1.0"},
		[]string{"application/json"},
		"",
	},
}

func TestContentNegotiation(t *testing.T) {
	for _, test := range contentNegotiationTestData {
		if val, _ := PreferredContentType(test.accept, test.available); val != test.expect {
			t.Errorf("[%v] Expected type %v got %v", test.name, test.expect, val)
		}
	}
}
