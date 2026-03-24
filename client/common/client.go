package common

import (
	"context"
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/bet"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
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
func (c *Client) StartClientLoop(ctx context.Context, bet *bet.Bet) {
	go func() {
		<-ctx.Done()
		log.Infof("action: graceful_shutdown | result: in_progress | client_id: %v", c.config.ID)
		if c.conn != nil {
			c.conn.Close()
		}
		log.Infof("action: graceful_shutdown | result: success | client_id: %v", c.config.ID)
	}()

	if err := c.createClientSocket(); err != nil {
		log.Criticalf("action: connect | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}
	defer c.conn.Close()

	if err := protocol.SendBet(c.conn, bet); err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	success, err := protocol.ReceiveAck(c.conn)
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	if success {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", bet.Document, bet.Number)
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | dni: %v | numero: %v", bet.Document, bet.Number)
	}
}
