package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

type Bet struct {
	Name         string
	Surname      string
	Document     string
	Birthdate    string
	Number       string
	BettingHouse string
}

func (b Bet) Encode() []byte {
	return []byte(fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		b.BettingHouse,
		b.Name,
		b.Surname,
		b.Document,
		b.Birthdate,
		b.Number,
	))
}

type BettingHouse struct {
	conn net.Conn
}

func BettingHouseConnect(addr string) (*BettingHouse, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &BettingHouse{conn: conn}, nil
}

func (b *BettingHouse) PlaceBets(bets []Bet, MaxAmountOfBets uint8) error {
	writer := bufio.NewWriter(b.conn)

	nbatches := (len(bets) + int(MaxAmountOfBets) - 1) / int(MaxAmountOfBets)
	log.Infof("Sending %d batches of bets", nbatches)
	err := binary.Write(writer, binary.BigEndian, uint32(nbatches))
	if err != nil {
		return fmt.Errorf("error writing bet length: %v", err)
	}

	// envio las apuestas agrupadas en paquetes de a MaxAmountOfBets
	for i := 0; i < len(bets); i += int(MaxAmountOfBets) {
		end := i + int(MaxAmountOfBets)
		if end > len(bets) {
			end = len(bets)
		}
		betsToSend := bets[i:end]

		betsEncoded := make([][]byte, len(betsToSend))
		for i, bet := range betsToSend {
			betsEncoded[i] = bet.Encode()
		}

		batchBytes := bytes.Join(betsEncoded, []byte("\n"))

		batchLen := len(batchBytes)
		batchLenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(batchLenBytes, uint32(batchLen))

		writer.Write(batchLenBytes)
		writer.Write(batchBytes)
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing bet: %v", err)
	}
	return nil
}

func (b *BettingHouse) Close() {
	b.conn.Close()
}
