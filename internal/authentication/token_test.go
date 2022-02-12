package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	minTokenLength = 14
)

func isGoodToken(assert *assert.Assertions, require *require.Assertions, token string, compat bool) {
	if compat {
		assert.Equal(len(token), compatTokenLength, "Unexpected compatibility token length")
	} else {
		assert.Equal(len(token), regularTokenLength, "Unexpected regular token length")
	}

	assert.GreaterOrEqual(len(token), minTokenLength, "Token is too short to give sufficient entropy")

	prefix := token[0:len(applicationTokenPrefix)]

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
