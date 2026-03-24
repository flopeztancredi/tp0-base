package protocol

import (
	"encoding/binary"
	"net"
)

type Decoder struct {
	conn net.Conn
	err  error
}

func NewDecoder(conn net.Conn) *Decoder {
	return &Decoder{conn: conn}
}

func (d *Decoder) ReadUint8() uint8 {
	if d.err != nil {
		return 0
	}
	buf, err := d.recvExact(1)
	if err != nil {
		d.err = err
		return 0
	}
	return buf[0]
}

func (d *Decoder) ReadUint16() uint16 {
	if d.err != nil {
		return 0
	}
	buf, err := d.recvExact(2)
	if err != nil {
		d.err = err
		return 0
	}
	return binary.BigEndian.Uint16(buf)
}

func (d *Decoder) ReadUint32() uint32 {
	if d.err != nil {
		return 0
	}
	buf, err := d.recvExact(4)
	if err != nil {
		d.err = err
		return 0
	}
	return binary.BigEndian.Uint32(buf)
}

func (d *Decoder) ReadString() string {
	length := d.ReadUint16()
	if d.err != nil {
		return ""
	}
	buf, err := d.recvExact(int(length))
	if err != nil {
		d.err = err
		return ""
	}
	return string(buf)
}

func (d *Decoder) ReadFixedString(length int) string {
	if d.err != nil {
		return ""
	}
	buf, err := d.recvExact(length)
	if err != nil {
		d.err = err
		return ""
	}
	return string(buf)
}

func (d *Decoder) Err() error {
	return d.err
}

func (d *Decoder) recvExact(n int) ([]byte, error) {
	buf := make([]byte, n)
	received := 0
	for received < n {
		r, err := d.conn.Read(buf[received:])
		if err != nil {
			return nil, err
		}
		received += r
	}
	return buf, nil
}
