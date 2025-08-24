package domain

import (
	"fmt"
)

type Bet struct {
	Nombre     string
	Apellido   string
	Documento  string
	Nacimiento string
	Numero     uint
}

// NewBet creates a new bet with validation
func NewBet(nombre, apellido, documento, nacimiento string, numero uint) (*Bet, error) {
	//TODO: Add validations
	
	return &Bet{
		Nombre:     nombre,
		Apellido:   apellido,
		Documento:  documento,
		Nacimiento: nacimiento,
		Numero:     numero,
	}, nil
}

// String representation of the bet
func (b *Bet) String() string {
	return fmt.Sprintf("Bet{DNI: %s, Number: %d, Name: %s %s, Birth: %s}", 
		b.Documento, b.Numero, b.Nombre, b.Apellido, b.Nacimiento)
}
