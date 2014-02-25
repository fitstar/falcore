package utils

import (
	"errors"
)

var ErrBufferFull = errors.New("Buffer Full")

type RingBuffer struct {
	buff []byte
	head int
	tail int
}

func NewRingBuffer(size int) *RingBuffer {
	// Always keep 1 byte empty to distinguish between full and empty
	// Make the internal buffer 1 byte larger than the requested size
	return &RingBuffer{make([]byte, size+1), 0, 0}
}

func (b *RingBuffer) Read(p []byte) (int, error) {
	var i int = 0

	if b.tail < b.head {
		// wraparound.  read to end
		i += copy(p, b.buff[b.head:])
		i += copy(p[i:], b.buff[0:b.tail])
	} else if b.tail > b.head {
		// read to tail
		i += copy(p, b.buff[b.head:b.tail])
	}

	b.head = (b.head + i) % len(b.buff)

	return i, nil
}

func (b *RingBuffer) Write(p []byte) (int, error) {
	var i int = 0

	// Write the stuff
	if b.tail >= b.head {
		// leave last cell open if head is at 0
		maxTail := len(b.buff)
		if b.head == 0 {
			maxTail -= 1
		}
		// Write to end of buff
		if b.tail < maxTail {
			i += copy(b.buff[b.tail:maxTail], p)
			b.tail += i
		}
		// Wrap around and write up to head
		if i < len(p) && b.head > 0 {
			b.tail = copy(b.buff[:b.head-1], p[i:])
			i += b.tail
		}
	} else {
		// write up to head
		i += copy(b.buff[b.tail:b.head-1], p)
		b.tail += i
	}

	if b.tail == len(b.buff) {
		b.tail = 0
	}

	var err error = nil
	if i < len(p) {
		err = ErrBufferFull
	}

	return i, err
}

func (b *RingBuffer) Free() int {
	return b.Cap() - b.Len()
}

func (b *RingBuffer) Cap() int {
	return cap(b.buff) - 1
}

func (b *RingBuffer) Len() int {
	if b.head < b.tail {
		return b.tail - b.head
	} else if b.tail < b.head {
		return len(b.buff) - (b.head - b.tail)
	}

	return 0
}

func (b *RingBuffer) Empty() bool {
	return b.head == b.tail
}

func (b *RingBuffer) Full() bool {
	return b.Cap() == b.Len()
}

func (b *RingBuffer) Reset() {
	b.head = 0
	b.tail = 0
}
