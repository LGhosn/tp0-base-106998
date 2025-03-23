package common

import (
	"bufio"
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

func (b *BettingHouse) PlaceBet(bet Bet) error {
	writer := bufio.NewWriter(b.conn)
	betEncoded := bet.Encode()
	betSize := uint32(len(betEncoded))
	err := binary.Write(writer, binary.BigEndian, betSize)
	if err != nil {
		return fmt.Errorf("error writing bet length: %v", err)
	}
	_, err = writer.Write(betEncoded)
	if err != nil {
		return fmt.Errorf("error writing bet: %v", err)
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
