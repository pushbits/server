package authentication

import (
	"crypto/rand"
	"math/big"
)

var (
	tokenCharacters   = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randomTokenLength = 64
	applicationPrefix = "A"
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
func GenerateNotExistingToken(generateToken func() string, tokenExists func(token string) bool) string {
	for {
		token := generateToken()

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

func generateRandomToken(prefix string) string {
	return prefix + generateRandomString(randomTokenLength)
}

// GenerateApplicationToken generates a token for an application.
func GenerateApplicationToken() string {
	return generateRandomToken(applicationPrefix)
}
