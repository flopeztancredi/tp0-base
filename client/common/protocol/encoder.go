package protocol

import (
	"bytes"
	"encoding/binary"
)

type Encoder struct {
	buf *bytes.Buffer
	err error
}

func NewEncoder() *Encoder {
	return &Encoder{
		buf: bytes.NewBuffer(nil),
	}
}

func (e *Encoder) WriteUint8(v uint8) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.buf, binary.BigEndian, v)
}

func (e *Encoder) WriteUint16(v uint16) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.buf, binary.BigEndian, v)
}

func (e *Encoder) WriteUint32(v uint32) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.buf, binary.BigEndian, v)
}

func (e *Encoder) WriteString(s string) {
	if e.err != nil {
		return
	}
	e.WriteUint16(uint16(len(s)))
	if e.err == nil {
		_, e.err = e.buf.WriteString(s)
	}
}

func (e *Encoder) WriteFixedString(s string) {
	if e.err != nil {
		return
	}
	_, e.err = e.buf.WriteString(s)
}

func (e *Encoder) Bytes() ([]byte, error) {
	return e.buf.Bytes(), e.err
}
