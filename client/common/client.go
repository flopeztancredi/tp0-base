package common

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/bet"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             string
	ServerAddress  string
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(ctx context.Context, r io.Reader) {
	go func() {
		<-ctx.Done()
		log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
		if c.conn != nil {
			c.conn.Close()
		}
		log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
	}()

	reader := csv.NewReader(r)
	batch := make([]*bet.Bet, 0, c.config.BatchMaxAmount)

	for {
		if ctx.Err() != nil {
			return
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_csv | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}

		b, err := bet.NewBetFromCSVRecord(c.config.ID, record)
		if err != nil {
			log.Errorf("action: parse_bet | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}
		batch = append(batch, b)

		if len(batch) >= c.config.BatchMaxAmount {
			if err := c.sendBatch(ctx, batch); err != nil {
				return
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := c.sendBatch(ctx, batch); err != nil {
			return
		}
	}

	if err := c.notifyAndQueryWinners(ctx); err != nil {
		return
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendBatch(ctx context.Context, bets []*bet.Bet) error {
	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: connect | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}
	defer c.conn.Close()

	if err := protocol.SendBatch(c.conn, bets); err != nil {
		if ctx.Err() != nil {
			return err
		}
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}

	success, err := protocol.ReceiveAck(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return err
		}
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}

	if !success {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %s | error: server returned failure", c.config.ID)
		return fmt.Errorf("server couldn't store bet")
	}
	log.Infof("action: apuesta_enviada | result: success | client_id: %s | cantidad: %d", c.config.ID, len(bets))
	return nil
}

func (c *Client) notifyAndQueryWinners(ctx context.Context) error {
	if err := c.createClientSocket(); err != nil {
		log.Errorf("action: connect | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}
	defer c.conn.Close()

	agencyID, _ := strconv.ParseUint(c.config.ID, 10, 32)
	if err := protocol.SendDone(c.conn, uint32(agencyID)); err != nil {
		if ctx.Err() != nil {
			return err
		}
		log.Errorf("action: notificar_fin | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}

	if err := protocol.SendWinnersRequest(c.conn, uint32(agencyID)); err != nil {
		if ctx.Err() != nil {
			return err
		}
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}

	winners, err := protocol.ReceiveWinnersResponse(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return err
		}
		log.Errorf("action: consulta_ganadores | result: fail | client_id: %s | error: %s", c.config.ID, err)
		return err
	}

	log.Infof("action: consulta_ganadores | result: success | client_id: %s | cant_ganadores: %d", c.config.ID, len(winners))
	return nil
}
