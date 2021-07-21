package credentials

import (
	"errors"
	"log"

	"github.com/alexedwards/argon2id"
)

// CreatePasswordHash returns a hashed version of the given password.
func (m *Manager) CreatePasswordHash(password string) ([]byte, error) {
	if m.checkHIBP {
		pwned, err := IsPasswordPwned(password)
		if err != nil {
			return []byte{}, errors.New("HIBP is not available, please wait until service is available again")
		} else if pwned {
			return []byte{}, errors.New("password is pwned, please choose another one")
		}
	}

	hash, err := argon2id.CreateHash(password, m.argon2Params)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return []byte(hash), nil
}

// ComparePassword compares a hashed password with its possible plaintext equivalent.
func ComparePassword(hash, password []byte) bool {
	match, err := argon2id.ComparePasswordAndHash(string(password), string(hash))

	if err != nil {
		log.Fatal(err)
		return false
	}

	return match
}
