package protocol

import (
	"fmt"
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/bet"
)

const (
	MsgTypeAck             uint8 = 0x02
	MsgTypeBatch           uint8 = 0x03
	MsgTypeDone            uint8 = 0x04
	MsgTypeWinnersRequest  uint8 = 0x05
	MsgTypeWinnersResponse uint8 = 0x06
	AckSuccess             uint8 = 0x00
	AckFailure             uint8 = 0x01
)

func SendBatch(conn net.Conn, bets []*bet.Bet) error {
	enc := NewEncoder()
	enc.WriteUint8(MsgTypeBatch)
	enc.WriteUint16(uint16(len(bets)))
	for _, bet := range bets {
		enc.WriteUint32(bet.Agency)
		enc.WriteString(bet.FirstName)
		enc.WriteString(bet.LastName)
		enc.WriteUint32(bet.Document)
		enc.WriteFixedString(bet.Birthdate)
		enc.WriteUint32(bet.Number)
	}

	data, err := enc.Bytes()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func SendDone(conn net.Conn, agency uint32) error {
	enc := NewEncoder()
	enc.WriteUint8(MsgTypeDone)
	enc.WriteUint32(agency)

	data, err := enc.Bytes()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func SendWinnersRequest(conn net.Conn, agency uint32) error {
	enc := NewEncoder()
	enc.WriteUint8(MsgTypeWinnersRequest)
	enc.WriteUint32(agency)

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

func ReceiveWinnersResponse(conn net.Conn) ([]uint32, error) {
	dec := NewDecoder(conn)
	msgType := dec.ReadUint8()
	if err := dec.Err(); err != nil {
		return nil, err
	}
	if msgType != MsgTypeWinnersResponse {
		return nil, fmt.Errorf("expected winners response message, got %d", msgType)
	}

	numWinners := dec.ReadUint16()
	if err := dec.Err(); err != nil {
		return nil, err
	}

	winners := make([]uint32, numWinners)
	for i := 0; i < int(numWinners); i++ {
		winners[i] = dec.ReadUint32()
	}
	if err := dec.Err(); err != nil {
		return nil, err
	}
	return winners, nil
}
