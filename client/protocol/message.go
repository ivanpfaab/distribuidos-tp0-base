package protocol

import (
	"fmt"
	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/domain"
)

// Message type constants
const (
	MessageTypeBatchBets     = 'B'
	MessageTypeNotification  = 'N'
	MessageTypeWinnerQuery   = 'W'
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
// <agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>&&
func (b *BetMessage) Format() string {
	betContent := fmt.Sprintf("%d|%s|%s|%s|%s|%d&&", 
		b.AgencyID, 
		b.Nombre, 
		b.Apellido, 
		b.Documento, 
		b.Nacimiento, 
		b.Numero)
	
	return betContent
}

// Formats batch bet message in the format: 
// <agency id>|<nombre>|<apellido>|<document>|<fecha nacimiento>|<numero>&&<agency id>...
func (b *BatchBetMessage) FormatBatch() string {
	msg := ""
	for _, bet := range b.Bets {
		betContent := bet.Format()
		msg += betContent
	}
	
	return msg
}

// NotificationMessage represents a completion notification
type NotificationMessage struct {
	AgencyID int
}

// NewNotificationMessage creates a new notification message
func NewNotificationMessage(agencyID int) *NotificationMessage {
	return &NotificationMessage{
		AgencyID: agencyID,
	}
}

// Format formats the notification message
func (n *NotificationMessage) Format() string {
	return fmt.Sprintf("%d", n.AgencyID)
}

// WinnerQueryMessage represents a query for winners from a specific agency
type WinnerQueryMessage struct {
	AgencyID int
}

// NewWinnerQueryMessage creates a new winner query message
func NewWinnerQueryMessage(agencyID int) *WinnerQueryMessage {
	return &WinnerQueryMessage{
		AgencyID: agencyID,
	}
}

// Format formats the winner query message
func (w *WinnerQueryMessage) Format() string {
	return fmt.Sprintf("%d", w.AgencyID)
}
