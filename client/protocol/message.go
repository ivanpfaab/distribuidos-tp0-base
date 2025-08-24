package protocol

import (
	"fmt"
)

// BetMessage represents a bet submission in simple string format
type BetMessage struct {
	MsgLength  uint8
	AgencyID   string
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     uint
}

// NewBetMessage creates a new bet message
func NewBetMessage(agencyID, nombre, apellido, documento, nacimiento string, numero uint) *BetMessage {
	msg := &BetMessage{
		AgencyID:   agencyID,
		Nombre:     nombre,
		Apellido:   apellido,
		Documento:  documento,
		Nacimiento: nacimiento,
		Numero:     numero,
	}
	
	// Calculate message length 
	msg.MsgLength = uint8(len(msg.String()))
	return msg
}

// String converts the bet message to the required format
func (b *BetMessage) String() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", 
		b.AgencyID, 
		b.Nombre, 
		b.Apellido, 
		b.Documento, 
		b.Nacimiento, 
		b.Numero)
}

// ParseResponse parses a simple response from the server
func ParseResponse(response string) (bool, error) {
	//TODO: when working with the server, change the logic to check if the bet was accepted or not
	if len(response) == 0 {
		return false, fmt.Errorf("empty response from server")
	} else {
		return true, nil
	}
}
