package utils

import (
	"bytes"
	"fmt"
	"testing"
)

// Apply these operations in order and verify state
var ringBufferTestData = []struct {
	name  string
	op    string
	size  int
	opLen int
	err   error
	exLen int
}{
	{"XX__", "w", 512, 512, nil, 512},
	{"_X__", "r", 256, 256, nil, 256},
	{"_XXX", "w", 512, 512, nil, 768},
	{"__XX", "r", 256, 256, nil, 512},
	{"X_XX", "w", 256, 256, nil, 768},
	{"XX|XX", "w", 256, 256, nil, 1024},
	{"__|__", "r", 1024, 1024, nil, 0},
	{"w16", "w", 16, 16, nil, 16},
	{"r16", "r", 16, 16, nil, 0},
	{"w16", "w", 16, 16, nil, 16},
	{"r32", "r", 32, 16, nil, 0},
	{"wfull", "w", 1024, 1024, nil, 1024},
	{"wfull", "w", 1024, 0, nil, 1024},
	{"empty", "r", 1024, 1024, nil, 0},
	{"wfill", "w", 768, 768, nil, 768},
	{"wfill", "w", 768, 256, nil, 1024},
	{"empty", "r", 1024, 1024, nil, 0},
	{"w1", "w", 1, 1, nil, 1},
	{"r1", "r", 1, 1, nil, 0},
}

func TestRingBuffer(t *testing.T) {
	b := NewRingBuffer(1024)
	bb := new(bytes.Buffer)

	if c := b.Cap(); c != 1024 {
		t.Errorf("Cap doesn't match expected 1024: %v", c)
	}

	for x := 0; x < 2048 && !t.Failed(); x++ {
		for testI, test := range ringBufferTestData {
			if test.op == "w" {
				wdata := make([]byte, test.size)
				for i := 0; i < test.size; i++ {
					wdata[i] = byte(testI)
				}
				i, err := b.Write(wdata)
				bb.Write(wdata[0:test.opLen])
				if i != test.opLen {
					t.Errorf("%v Write len expected %v, got %v", test.name, test.opLen, i)
				}
				if err != test.err {
					t.Errorf("%v Unexpected error writing: %v, expected %v", test.name, err, test.err)
				}
				if l := b.Len(); l != test.exLen {
					t.Errorf("%v Post-write len expected %v, got %v", test.name, test.exLen, l)
				}
				if f := b.Free(); f != 1024-test.exLen {
					t.Errorf("%v Post-write free expected %v, got %v", test.name, 1024-test.exLen, f)
				}
				if testI == 5 {
					// t.Errorf("%x", b.buff)
				}

			} else {
				rdata := make([]byte, test.size)
				expectData := make([]byte, test.opLen)
				i, err := b.Read(rdata)
				bb.Read(expectData)
				if i != test.opLen {
					t.Errorf("%v Read len expected %v, got %v", test.name, test.opLen, i)
				}
				if err != test.err {
					t.Errorf("%v Unexpected error reading: %v, expected %v", test.name, err, test.err)
				}
				if l := b.Len(); l != test.exLen {
					t.Errorf("%v Post-read len expected %v, got %v", test.name, test.exLen, l)
				}
				if f := b.Free(); f != 1024-test.exLen {
					t.Errorf("%v Post-read free expected %v, got %v", test.name, 1024-test.exLen, f)
				}
				if !bytes.Equal(rdata[0:i], expectData) {
					t.Errorf("%v Data read doesn't match expectation\nEXPECTED: %x\nACTUAL:   %x", test.name, expectData, rdata[0:i])
				}
			}

			if t.Failed() {
				return
			}
		}

		fmt.Println(x, "==================================")
	}

}
