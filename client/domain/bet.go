package domain

import (
	"fmt"
)

type Bet struct {
	AgencyID   int
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     uint
}

// NewBet creates a new bet with validation
func NewBet(agencyID int, nombre string, apellido string, documento string, nacimiento string, numero uint) (*Bet, error) {
	//TODO: Add validations
	
	return &Bet{
		AgencyID:   agencyID,
		Nombre:     nombre,
		Apellido:   apellido,
		Documento:  documento,
		Nacimiento: nacimiento,
		Numero:     numero,
	}, nil
}

// String representation of the bet
func (b *Bet) String() string {
	return fmt.Sprintf("Bet{AgencyID: %d, DNI: %s, Number: %d, Name: %s %s, Birth: %s}", 
		b.AgencyID, b.Documento, b.Numero, b.Nombre, b.Apellido, b.Nacimiento)
}

// BetsFromChunk creates a new bet from a chunk
func BetsFromChunk(chunk [][]string) []*Bet {
	var bets []*Bet
	for _, row := range chunk {
		bets = append(bets, NewBet(row[0], row[1], row[2], row[3], row[4], row[5]))
	}
	return bets
}
