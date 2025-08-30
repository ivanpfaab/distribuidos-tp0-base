package protocol

import (
	"fmt"
)

// BetMessage represents a bet submission in simple string format
type BetMessage struct {
	MsgLength  uint8
	AgencyID   int
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     uint
}

type BatchBetMessage struct {
	MsgSize uint8
	Bets    []*BetMessage
}

// NewBetMessage creates a new bet message
func NewBetMessage(agencyID int, nombre, apellido, documento, nacimiento string, numero uint) *BetMessage {
	msg := &BetMessage{
		AgencyID:   agencyID,
		Nombre:     nombre,
		Apellido:   apellido,
		Documento:  documento,
		Nacimiento: nacimiento,
		Numero:     numero,
	}
	
	// Calculate message length in bytes
	msg.MsgLength = msg.msgLength()+1
	return msg
}

// NewBatchBetMessage creates a new batch bet message
func NewBatchBetMessage(bets []*domain.Bet) *BatchBetMessage {
	betMessages := make([]*BetMessage, 0, len(bets))
	msgSize := 0
	for i, bet := range bets {
		betMessage := NewBetMessage(bet.AgencyID, bet.Nombre, bet.Apellido, bet.Documento, bet.Nacimiento, bet.Numero)
		betMessages = append(betMessages, betMessage)
		msgSize += betMessage.MsgLength
	}
	return &BatchBetMessage{
		MsgSize: msgSize+1,
		Bets: betMessages,
	}
}

func (b *BetMessage) msgLength() uint8 {
	msg := fmt.Sprintf("%d|%s|%s|%s|%s|%d", 
	b.AgencyID, 
	b.Nombre, 
	b.Apellido, 
	b.Documento, 
	b.Nacimiento, 
	b.Numero)

	return uint8(len(msg))
}

// converts a bet message to the format:
// <msg length><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
func (b *BetMessage) Format() string {
	return fmt.Sprintf("%d%d|%s|%s|%s|%s|%d", 
		b.MsgLength,
		b.AgencyID, 
		b.Nombre, 
		b.Apellido, 
		b.Documento, 
		b.Nacimiento, 
		b.Numero)
}

// Formats batch bet message in the format: 
// <msg size>
// <bet size 1><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
// <bet size 2><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
// ...
// <bet size n><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
func (b *BatchBetMessage) FormatBatch() string {
	msg := fmt.Sprintf("%d", b.MsgSize)
	for _, bet := range b.Bets {
		msg += bet.Format()
	}
	return msg
}
