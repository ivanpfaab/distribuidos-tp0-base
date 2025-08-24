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
	
	// Calculate message length 
	msg.MsgLength = uint8(len(msg.String()))
	return msg
}

// String converts the bet message to the required format
func (b *BetMessage) String() string {
	return fmt.Sprintf("%d|%s|%s|%s|%s|%d", 
		b.AgencyID, 
		b.Nombre, 
		b.Apellido, 
		b.Documento, 
		b.Nacimiento, 
		b.Numero)
}

// ParseResponse parses a simple response from the server
func ParseResponse(response string) (bool, error) {
	if len(response) == 0 {
		return false, fmt.Errorf("empty response from server")
	}
	
	// Check if response is a number (bet number as ACK)
	if _, err := fmt.Sscanf(response, "%d", new(int)); err == nil {
		return true, nil
	} else {
		return false, fmt.Errorf("server response: %s", response)
	}
}
