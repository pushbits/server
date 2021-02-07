package authentication

import (
	"crypto/rand"
	"log"
	"math/big"
)

var (
	tokenCharacters     = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	standardTokenLength = 64 // This length includes the prefix (one character).
	compatTokenLength   = 15 // This length includes the prefix (one character).
	applicationPrefix   = "A"
)

func randIntn(n int) int {
	max := big.NewInt(int64(n))

	res, err := rand.Int(rand.Reader, max)
	if err != nil {
		panic("random source is not available")
	}

	return int(res.Int64())
}

// GenerateNotExistingToken receives a token generation function and a function to check whether the token exists, returns a unique token.
func GenerateNotExistingToken(generateToken func(bool) string, compat bool, tokenExists func(token string) bool) string {
	for {
		token := generateToken(compat)

		if !tokenExists(token) {
			return token
		}
	}
}

func generateRandomString(length int) string {
	res := make([]byte, length)

	for i := range res {
		index := randIntn(len(tokenCharacters))
		res[i] = tokenCharacters[index]
	}

	return string(res)
}

func generateRandomToken(prefix string, compat bool) string {
	tokenLength := standardTokenLength

	if compat {
		tokenLength = compatTokenLength
	}

	// Although constant at the time of writing, this check should prevent future changes from generating insecure tokens.
	randomLength := tokenLength - len(prefix)
	if randomLength < 14 {
		log.Fatalf("Tokens should have more than %d random characters", randomLength)
	}

	return prefix + generateRandomString(randomLength)
}

// GenerateApplicationToken generates a token for an application.
func GenerateApplicationToken(compat bool) string {
	return generateRandomToken(applicationPrefix, compat)
}
