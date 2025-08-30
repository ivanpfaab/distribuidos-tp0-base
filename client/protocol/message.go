package protocol

import (
	"fmt"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/domain"
)

// BetMessage represents a bet submission in simple string format
type BetMessage struct {
	AgencyID   int
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     uint
}

type BatchBetMessage struct {
	Bets []*BetMessage
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
	
	// MsgLength is no longer needed since we calculate bet size dynamically in Format()
	return msg
}

// NewBatchBetMessage creates a new batch bet message
func NewBatchBetMessage(bets []*domain.Bet) *BatchBetMessage {
	betMessages := make([]*BetMessage, 0, len(bets))
	for _, bet := range bets {
		betMessage := NewBetMessage(bet.AgencyID, bet.Nombre, bet.Apellido, bet.Documento, bet.Nacimiento, bet.Numero)
		betMessages = append(betMessages, betMessage)
	}
	return &BatchBetMessage{
		Bets: betMessages,
	}
}

// converts a bet message to the format:
// <bet size><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>
func (b *BetMessage) Format() string {
	// Calculate the actual bet message content (without the size prefix)
	betContent := fmt.Sprintf("%d|%s|%s|%s|%s|%d", 
		b.AgencyID, 
		b.Nombre, 
		b.Apellido, 
		b.Documento, 
		b.Nacimiento, 
		b.Numero)
	
	// The bet size is the length of the bet content
	betSize := len(betContent)
	
	// Format bet size as a 2-digit string (00-99)
	// This allows us to handle bet sizes up to 99 characters
	sizeStr := fmt.Sprintf("%02d", betSize)
	
	return fmt.Sprintf("%s%s", sizeStr, betContent)
}

// Formats batch bet message in the format: 
// <msg size (8 bytes)><bet size 1><agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero><bet size 2>...
func (b *BatchBetMessage) FormatBatch() string {
	// Calculate total message size (sum of all bet message lengths)
	var totalSize uint32
	for _, bet := range b.Bets {
		betContent := fmt.Sprintf("%d|%s|%s|%s|%s|%d", 
			bet.AgencyID, 
			bet.Nombre, 
			bet.Apellido, 
			bet.Documento, 
			bet.Nacimiento, 
			bet.Numero)
		totalSize += uint32(len(betContent))
	}
	
	// Format message size as 8-byte string (00000000-99999999)
	sizeStr := fmt.Sprintf("%08d", totalSize)
	
	msg := sizeStr
	for _, bet := range b.Bets {
		msg += bet.Format()
	}
	return msg
}
