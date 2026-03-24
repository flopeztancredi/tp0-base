package protocol

import (
	"fmt"
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/common"
)

const (
	MsgTypeBet uint8 = 0x01
	MsgTypeAck uint8 = 0x02
	AckSuccess uint8 = 0x00
	AckFailure uint8 = 0x01
)

func recvExact(conn net.Conn, n int) ([]byte, error) {
	buf := make([]byte, n)
	received := 0
	for received < n {
		r, err := conn.Read(buf[received:])
		if err != nil {
			return nil, err
		}
		received += r
	}
	return buf, nil
}

func SendBet(conn net.Conn, bet *common.Bet) error {
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
	msgTypeBuf, err := recvExact(conn, 1)
	if err != nil {
		return false, err
	} else if msgTypeBuf[0] != MsgTypeAck {
		return false, fmt.Errorf("expected message type %v but received %v", MsgTypeAck, msgTypeBuf[0])
	}

	statusBuf, err := recvExact(conn, 1)
	if err != nil {
		return false, err
	}

	return statusBuf[0] == AckSuccess, nil
}
