package api

import (
	"testing"

	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
)

func TestApi_Health(t *testing.T) {
	assert := assert.New(t)
	handler := HealthHandler{
		DB: TestDatabase,
	}

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "-", Method: "GET", Endpoint: "/health", Data: "", ShouldStatus: 200})

	for _, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}
		handler.Health(c)

		assert.Equalf(w.Code, req.ShouldStatus, "Health should result in status code %d but code is %d", req.ShouldStatus, w.Code)
	}
}
