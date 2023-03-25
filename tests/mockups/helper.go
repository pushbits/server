package mockups

import (
	"crypto/rand"
	"encoding/base64"
)

func randStr(len int) string {
	buff := make([]byte, len)

	_, err := rand.Read(buff)
	if err != nil {
		panic("cannot read random data")
	}

	str := base64.StdEncoding.EncodeToString(buff)

	// Base 64 can be longer than len
	return str[:len]
}
