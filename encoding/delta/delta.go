package delta

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/segmentio/parquet-go/internal/unsafecast"
)

type int32Buffer struct {
	values []int32
}

func (buf *int32Buffer) resize(size int) {
	if cap(buf.values) < size {
		buf.values = make([]int32, size, 2*size)
	} else {
		buf.values = buf.values[:size]
	}
}

func (buf *int32Buffer) decode(src []byte) ([]byte, error) {
	values, remain, err := decodeInt32(unsafecast.Int32ToBytes(buf.values[:0]), src)
	buf.values = unsafecast.BytesToInt32(values)
	return remain, err
}

var (
	int32BufferPool sync.Pool // *int32Buffer
)

func getInt32Buffer() *int32Buffer {
	b, _ := int32BufferPool.Get().(*int32Buffer)
	if b != nil {
		b.values = b.values[:0]
	} else {
		b = &int32Buffer{
			values: make([]int32, 0, 1024),
		}
	}
	return b
}

func putInt32Buffer(b *int32Buffer) {
	int32BufferPool.Put(b)
}

func bytesToInt32(b []byte) []int32 {
	return unsafe.Slice(*(**int32)(unsafe.Pointer(&b)), len(b)/4)
}

func bytesToInt64(b []byte) []int64 {
	return unsafe.Slice(*(**int64)(unsafe.Pointer(&b)), len(b)/8)
}

func resizeNoMemclr(buf []byte, size int) []byte {
	if cap(buf) < size {
		return grow(buf, size)
	}
	return buf[:size]
}

func resize(buf []byte, size int) []byte {
	if cap(buf) < size {
		return grow(buf, size)
	}
	if size > len(buf) {
		clear := buf[len(buf):size]
		for i := range clear {
			clear[i] = 0
		}
	}
	return buf[:size]
}

func grow(buf []byte, size int) []byte {
	newCap := 2 * cap(buf)
	if newCap < size {
		newCap = size
	}
	newBuf := make([]byte, size, newCap)
	copy(newBuf, buf)
	return newBuf
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func errPrefixAndSuffixLengthMismatch(prefixLength, suffixLength int) error {
	return fmt.Errorf("length of prefix and suffix mismatch: %d != %d", prefixLength, suffixLength)
}

func errInvalidNegativeValueLength(length int) error {
	return fmt.Errorf("invalid negative value length: %d", length)
}

func errInvalidNegativePrefixLength(length int) error {
	return fmt.Errorf("invalid negative prefix length: %d", length)
}

func errValueLengthOutOfBounds(length, maxLength int) error {
	return fmt.Errorf("value length is larger than the input size: %d > %d", length, maxLength)
}

func errPrefixLengthOutOfBounds(length, maxLength int) error {
	return fmt.Errorf("prefix length %d is larger than the last value of size %d", length, maxLength)
}
