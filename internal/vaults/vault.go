package vaults

import "strings"

// Vault is the basic skeleton of a vault implementation.
type Vault interface{ Entries() []Entry }

// Type is an implemented vault type name.
type Type string

func (t Type) String() string { return string(t) }

const (
	ANDOTP  Type = "andotp"
	AEGIS   Type = "aegis"
	TWOFAS  Type = "twofas"
	STRATUM Type = "stratum"
	KEEPASS Type = "keepass"
	PROTON  Type = "proton"
)

// Returns a list containing the implemented types.
func Types() []Type {
	return []Type{
		ANDOTP,
		AEGIS,
		TWOFAS,
		STRATUM,
		KEEPASS,
		PROTON,
	}
}

// StrTypes returns a concatenated string of all defined types.
func StrTypes() string {
	var s []string
	for _, t := range Types() {
		s = append(s, t.String())
	}
	return strings.Join(s, ", ")
}
