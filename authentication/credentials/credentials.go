package credentials

import (
	"log"

	"github.com/pushbits/server/configuration"

	"github.com/alexedwards/argon2id"
)

// Manager holds information for managing credentials.
type Manager struct {
	checkHIBP    bool
	argon2Params *argon2id.Params
}

// CreateManager instanciates a credential manager.
func CreateManager(checkHIBP bool, c configuration.CryptoConfig) *Manager {
	log.Println("Setting up credential manager.")

	argon2Params := &argon2id.Params{
		Memory:      c.Argon2.Memory,
		Iterations:  c.Argon2.Iterations,
		Parallelism: c.Argon2.Parallelism,
		SaltLength:  c.Argon2.SaltLength,
		KeyLength:   c.Argon2.KeyLength,
	}

	return &Manager{
		checkHIBP:    checkHIBP,
		argon2Params: argon2Params,
	}
}
