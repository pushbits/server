package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
)

func TestApi_getID(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)
	testValue := uint(1337)

	testCases := make(map[interface{}]tests.Request)
	testCases[-1] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}
	testCases[uint(1)] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[uint(0)] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[uint(500)] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[500] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}
	testCases["test"] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}
	testCases[model.Application{}] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}
	testCases[&model.Application{}] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}
	testCases[&testValue] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 500}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("id", id)
		getID(c)

		assert.Equalf(w.Code, req.ShouldStatus, "getApi id was set to %v (%T) and should result in status code %d but code is %d", id, id, req.ShouldStatus, w.Code)
	}
}
