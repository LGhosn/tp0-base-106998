package common

import (
	"os"
	"os/signal"
	"syscall"

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
	conn   *BettingHouse
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
	conn, err := BettingHouseConnect(c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// HandleSignals Handle SIGINT and SIGTERM signals
func (c *Client) HandleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Infof("action: signal_received | result: success | client_id: %v | signal: %v",
			c.config.ID,
			sig,
		)
		os.Exit(0)
	}()
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// Handle SIGINT and SIGTERM
	c.HandleSignals()

	log.Infof("action: pre_socket")

	// Create the connection the server.
	c.createClientSocket()
	defer c.conn.Close()

	log.Infof("action: connect | result: success | client_id: %v", c.config.ID)

	bet := Bet{
		Name:         os.Getenv("NOMBRE"),
		Surname:      os.Getenv("APELLIDO"),
		Document:     os.Getenv("DOCUMENTO"),
		Birthdate:    os.Getenv("FECHA_NACIMIENTO"),
		Number:       os.Getenv("NUMERO"),
		BettingHouse: c.config.ID,
	}
	log.Infof(
		"action: apuesta_generada | result: success | dni: %v | numero: %v",
		bet.Document,
		bet.Number,
	)

	err := c.conn.PlaceBet(bet)
	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | error: %v", err)
	} else {
		log.Infof(
			"action: apuesta_enviada | result: success | dni: %v | numero: %v",
			bet.Document,
			bet.Number,
		)
	}

}
