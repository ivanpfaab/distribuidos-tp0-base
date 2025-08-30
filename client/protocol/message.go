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
	msg.MsgLength = msg.msgLength()
	return msg
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
