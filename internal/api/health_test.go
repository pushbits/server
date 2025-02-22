package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pushbits/server/tests"
)

func TestApi_Health(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)
	handler := HealthHandler{
		DB: ctx.Database,
	}

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "-", Method: "GET", Endpoint: "/health", Data: "", ShouldStatus: 200})

	for _, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatal(err.Error())
		}
		handler.Health(c)

		assert.Equalf(w.Code, req.ShouldStatus, "Health should result in status code %d but code is %d", req.ShouldStatus, w.Code)
	}
}
