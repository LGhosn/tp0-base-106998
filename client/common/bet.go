package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

const BET_FLAG = 1
const END_FLAG = 2

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
	conn     net.Conn
	agencyId uint8
}

func (b *BettingHouse) AllBetsSent() error {
	writer := bufio.NewWriter(b.conn)
	err := binary.Write(writer, binary.BigEndian, uint8(END_FLAG))
	log.Infof("All bets sent flag %d agency %d", uint8(END_FLAG), b.agencyId)
	if err != nil {
		return fmt.Errorf("error writing bet length: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing bet: %v", err)
	}
	return nil
}

func (b *BettingHouse) GetWinners() ([]string, error) {
	// Leer el número de DNIs
	amountOfWinners := make([]byte, 4)
	if _, err := io.ReadFull(b.conn, amountOfWinners); err != nil {
		return nil, fmt.Errorf("couldn't recv winner quantity: %v", err)
	}

	amount := binary.BigEndian.Uint32(amountOfWinners)

	// No hay ganadores, devuelvo lista vacía
	if amount == 0 {
		return nil, nil
	}

	// Leeo los DNIs
	winners := make([]byte, 4*amount)
	if _, err := io.ReadFull(b.conn, winners); err != nil {
		return nil, fmt.Errorf("couldn't recv winners: %v", err)
	}

	// armo la lista de ganadores en formato string
	winnersStr := make([]string, amount)
	for i := 0; i < int(amount); i++ {
		winner := binary.BigEndian.Uint32(winners[i*4 : (i+1)*4])
		winnersStr[i] = fmt.Sprintf("%d", winner)
	}

	return winnersStr, nil
}

func BettingHouseConnect(addr string, agencyId string) (*BettingHouse, error) {
	id, err := strconv.ParseUint(agencyId, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("error parsing agency id: %v", err)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriter(conn)
	err = binary.Write(writer, binary.BigEndian, uint8(id))
	if err != nil {
		return nil, fmt.Errorf("error writing agency id: %v", err)
	}
	err = writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("error flushing agency id: %v", err)
	}
	return &BettingHouse{conn: conn, agencyId: uint8(id)}, nil
}

func (b *BettingHouse) PlaceBets(bets []Bet, MaxAmountOfBets uint8) error {
	writer := bufio.NewWriter(b.conn)

	err := binary.Write(writer, binary.BigEndian, uint8(BET_FLAG))
	if err != nil {
		return fmt.Errorf("error writing bet length: %v", err)
	}

	nbatches := (len(bets) + int(MaxAmountOfBets) - 1) / int(MaxAmountOfBets)
	err = binary.Write(writer, binary.BigEndian, uint32(nbatches))
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
	if b.conn != nil {
		b.conn.Close()
	}
}
