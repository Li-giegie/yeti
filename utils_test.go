package yeti

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"testing"
)

func TestToStr(t *testing.T) {
	for i, a := range []any{
		1, int8(2), int16(3), int32(4), int64(5),
		uint(1), uint8(2), uint16(3), uint32(4), uint64(5),
		float32(3.1415926), float64(1.234567890),
		false,
		New(1), New(int8(2)), New(int16(3)), New(int32(4)), New(int64(5)),
		New(uint(1)), New(uint8(2)), New(uint16(3)), New(uint32(4)), New(uint64(5)),
		New(float32(3.1415926)), New(float64(1.234567890)),
		New(false), (*bool)(nil), New(interface{}(nil)),
		New([]string{"", "123", "abc"}),
		struct {
			I int
			F float64
			B bool
		}{I: 1,
			F: 3.1415926,
			B: true},
	} {
		ret, err := toString(&a)
		if err != nil {
			t.Errorf("#%d: toStr err: %v", i, err)
			continue
		}
		t.Logf("#%d: %#v", i, ret)
	}
}

func New[T any](a T) *T {
	return &a
}

func TestName(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buf)

	mw.WriteField("key", "value1")
	mw.WriteField("key", "value2")
	mw.WriteField("key", "value3")
	fmt.Println(buf.String())
}
