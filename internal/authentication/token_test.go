package authentication

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isGoodToken(assert *assert.Assertions, require *require.Assertions, token string, compat bool) {
	prefix := token[0:len(applicationTokenPrefix)]
	token = token[len(applicationTokenPrefix):]

	// Although constant at the time of writing, this check should prevent future changes from generating insecure tokens.
	if len(token) < 14 {
		log.Fatalf("Tokens should have more random characters")
	}

	if compat {
		assert.Equal(len(token), compatTokenLength, "Unexpected compatibility token length")
	} else {
		assert.Equal(len(token), regularTokenLength, "Unexpected regular token length")
	}

	assert.Equal(prefix, applicationTokenPrefix, "Invalid token prefix")

	for _, c := range []byte(token) {
		assert.Contains(tokenCharacters, c, "Unexpected character in token")
	}
}

func TestAuthentication_GenerateApplicationToken(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for i := 0; i < 64; i++ {
		token := GenerateApplicationToken(false)

		isGoodToken(assert, require, token, false)
	}

	for i := 0; i < 64; i++ {
		token := GenerateApplicationToken(true)

		isGoodToken(assert, require, token, true)
	}
}
