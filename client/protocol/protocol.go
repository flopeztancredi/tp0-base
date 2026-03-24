package protocol

import (
	"fmt"
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/bet"
)

const (
	MsgTypeBet uint8 = 0x01
	MsgTypeAck uint8 = 0x02
	AckSuccess uint8 = 0x00
	AckFailure uint8 = 0x01
)

func SendBet(conn net.Conn, bet *bet.Bet) error {
	enc := NewEncoder()
	enc.WriteUint8(MsgTypeBet)
	enc.WriteUint32(bet.Agency)
	enc.WriteString(bet.FirstName)
	enc.WriteString(bet.LastName)
	enc.WriteUint32(bet.Document)
	enc.WriteFixedString(bet.Birthdate)
	enc.WriteUint32(bet.Number)

	data, err := enc.Bytes()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func ReceiveAck(conn net.Conn) (bool, error) {
	dec := NewDecoder(conn)
	msgType := dec.ReadUint8()
	if err := dec.Err(); err != nil {
		return false, err
	}
	if msgType != MsgTypeAck {
		return false, fmt.Errorf("expected ack message, got %d", msgType)
	}

	status := dec.ReadUint8()
	if err := dec.Err(); err != nil {
		return false, err
	}
	return status == AckSuccess, nil
}
