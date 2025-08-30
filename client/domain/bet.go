package domain

import (
	"fmt"
	"strconv"
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
func BetsFromChunk(chunk [][]string, agencyID int) ([]*Bet, error) {
	var bets []*Bet
	for _, row := range chunk {
		// Convert string values to appropriate types
		// CSV format: Nombre,Apellido,Documento,Nacimiento,Numero
		// We need to map: row[0]=Nombre, row[1]=Apellido, row[2]=Documento, row[3]=Nacimiento, row[4]=Numero
		
		numero, err := strconv.ParseUint(row[4], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid numero: %s", row[4])
		}
		
		bet, err := NewBet(agencyID, row[0], row[1], row[2], row[3], uint(numero))
		if err != nil {
			return nil, fmt.Errorf("failed to create bet: %w", err)
		}
		
		bets = append(bets, bet)
	}
	return bets, nil
}
