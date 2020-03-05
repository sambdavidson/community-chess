package tlsconsts

// Organization of Community Chessa
func Organization() []string {
	return []string{"Community Chess"}
}

// Country of Community Chessa
func Country() []string {
	return []string{"US"}
}

// Province of Community Chessa
func Province() []string {
	return []string{""}
}

// Locality of Community Chessa
func Locality() []string {
	return []string{"Seattle, WA"}
}

// StreetAddress of Community Chessa
func StreetAddress() []string {
	return []string{""}
}

// PostalCode of Community Chessa
func PostalCode() []string {
	return []string{""}
}

// SAN for x509 cert Subject Alternative Names
type SAN string

func (s SAN) String() string {
	return string(s)
}

// SAN constants
const (
	GameServer      SAN = "gameserver"
	GameMaster      SAN = "gamemaster"
	GameSlave       SAN = "gameslave"
	PlayerRegistrar SAN = "playerregistrar"
	Admin           SAN = "admin"
	Internal        SAN = "internal"
)
