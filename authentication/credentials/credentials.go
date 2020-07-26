package credentials

import "golang.org/x/crypto/bcrypt"

// CreatePassword returns a hashed version of the given password.
func CreatePassword(pw string) []byte {
	strength := 12
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), strength)

	if err != nil {
		panic(err)
	}

	return hashedPassword
}

// ComparePassword compares a hashed password with its possible plaintext equivalent.
func ComparePassword(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}
