package common

import (
	"fmt"
	"time"
)

type Bet struct {
	Agency    uint32
	FirstName string
	LastName  string
	Document  uint32
	Birthdate string
	Number    uint32
}

func NewBet(agency uint32, firstName string, lastName string, document uint32, birthdate string, number uint32) (*Bet, error) {
	bet := &Bet{
		Agency:    agency,
		FirstName: firstName,
		LastName:  lastName,
		Document:  document,
		Birthdate: birthdate,
		Number:    number,
	}

	if err := bet.validate(); err != nil {
		return nil, err
	}

	return bet, nil
}

func (b *Bet) validate() error {
	if b.Agency == 0 {
		return fmt.Errorf("agency must be greater than 0")
	}
	if b.FirstName == "" {
		return fmt.Errorf("first name must not be empty")
	}
	if b.LastName == "" {
		return fmt.Errorf("last name must not be empty")
	}
	if b.Document == 0 {
		return fmt.Errorf("document must be greater than 0")
	}
	if _, err := time.Parse("2006-01-02", b.Birthdate); err != nil {
		return fmt.Errorf("birthday must be in the format YYYY-MM-DD")
	}
	if b.Number == 0 {
		return fmt.Errorf("number must be greater than 0")
	}
	return nil
}
