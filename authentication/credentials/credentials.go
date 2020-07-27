package credentials

import (
	"log"

	"github.com/alexedwards/argon2id"
)

// CreatePasswordHash returns a hashed version of the given password.
func CreatePasswordHash(password string) []byte {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		panic(err)
	}

	return []byte(hash)
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
