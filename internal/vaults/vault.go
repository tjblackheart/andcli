package vaults

// Vault is the basic skeleton of a vault implementation.
type Vault interface {
	Entries() []Entry
}

const (
	TYPE_ANDOTP  string = "andotp"
	TYPE_AEGIS   string = "aegis"
	TYPE_TWOFAS  string = "twofas"
	TYPE_STRATUM string = "stratum"
	TYPE_KEEPASS string = "keepass"
)

// Returns a list containing the implemented types.
func Types() []string {
	return []string{
		TYPE_ANDOTP,
		TYPE_AEGIS,
		TYPE_TWOFAS,
		TYPE_STRATUM,
		TYPE_KEEPASS,
	}
}
