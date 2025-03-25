package common

import (
	"encoding/csv"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

const BETS_FILE = "/bets.csv"

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID              string
	ServerAddress   string
	MaxAmountOfBets string
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
	conn, err := BettingHouseConnect(c.config.ServerAddress, c.config.ID)
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

	betFile, err := os.Open(BETS_FILE)
	if err != nil {
		log.Criticalf("action: open_file | result: fail | error: %v", err)
	}
	defer betFile.Close()

	// Create the connection the server.
	c.createClientSocket()
	defer c.conn.Close()

	betReader := csv.NewReader(betFile)
	bets := make([]Bet, 0)

	// Read the file and send the bets
	for {
		readed, err := betReader.Read()
		if err != nil {
			break
		}

		bets = append(bets, Bet{
			BettingHouse: c.config.ID,
			Name:         readed[0],
			Surname:      readed[1],
			Document:     readed[2],
			Birthdate:    readed[3],
			Number:       readed[4],
		})

	}

	maxBets, parseErr := strconv.ParseUint(c.config.MaxAmountOfBets, 10, 8)
	if parseErr != nil {
		log.Criticalf("action: parse_max_bets | result: fail | error: %v", parseErr)
		return
	}
	err = c.conn.PlaceBets(bets, uint8(maxBets))
	if err != nil {
		log.Errorf("action: apuestas_enviadas | result: fail | client_id: %v | error: %v", err, c.config.ID)
	} else {
		log.Infof("action: apuestas_enviadas | result: success | client_id: %v", c.config.ID)
	}

	time.Sleep(5 * time.Second)

	err = c.conn.AllBetsSent()
	if err != nil {
		log.Errorf("action: todas_las_apuestas_enviadas | result: fail | client_id: %v | error: %v", err, c.config.ID)
	} else {
		log.Infof("action: todas_las_apuestas_enviadas | result: success | client_id: %v", c.config.ID)
	}

	time.Sleep(5 * time.Second)

	winners, err := c.conn.GetWinners()
	if err != nil {
		log.Errorf("action: consulta_ganadores  | result: fail | client_id: %v | error: %v", err, c.config.ID)
	} else {
		log.Infof("action: consulta_ganadores  | result: success | client_id: %v | cant_ganadores: %v", c.config.ID, len(winners))
	}

}
