package model

import (
	"time"

	"github.com/fsvxavier/nexs-lib/uid/interfaces"
)

// IDData é a implementação concreta de interfaces.IDData
type IDData struct {
	Timestamp  time.Time
	Value      string
	HexValue   string
	UUIDString string
	HexBytes   []byte
	Type       interfaces.IDType
}

// ToInterface converte o model.IDData para interfaces.IDData
func (d *IDData) ToInterface() *interfaces.IDData {
	return &interfaces.IDData{
		Timestamp:  d.Timestamp,
		Value:      d.Value,
		HexValue:   d.HexValue,
		UUIDString: d.UUIDString,
		HexBytes:   d.HexBytes,
		Type:       d.Type,
	}
}

// FromInterface converte interfaces.IDData para model.IDData
func FromInterface(d *interfaces.IDData) *IDData {
	return &IDData{
		Timestamp:  d.Timestamp,
		Value:      d.Value,
		HexValue:   d.HexValue,
		UUIDString: d.UUIDString,
		HexBytes:   d.HexBytes,
		Type:       d.Type,
	}
}
