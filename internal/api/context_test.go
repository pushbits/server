package api

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/pushbits/server/tests/mockups"
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
		idReturned, err := getID(c)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			idUint, ok := id.(uint)
			if ok {
				assert.Equalf(idReturned, idUint, "getApi id was set to %d but result is %d", idUint, idReturned)
			}
			assert.NoErrorf(err, "getId with id %v (%t) returned an error altough it should not: %v", id, id, err)
		} else {
			assert.Errorf(err, "getId with id %v (%t) returned no error altough it should", id, id)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "getApi id was set to %v (%T) and should result in status code %d but code is %d", id, id, req.ShouldStatus, w.Code)
	}
}

func TestApi_getApplication(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	applications := mockups.GetAllApplications()
	mockups.AddApplicationsToDb(TestDatabase, applications)

	// No testing of invalid ids as that is tested in TestApi_getID already
	testCases := make(map[uint]tests.Request)
	testCases[500] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 404}
	testCases[1] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[2] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("id", id)
		app, err := getApplication(c, TestDatabase)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			assert.Equalf(app.ID, id, "getApplication id was set to %d but resulting app id is %d", id, app.ID)
			assert.NoErrorf(err, "getApplication with id %v (%t) returned an error altough it should not: %v", id, id, err)
		} else {
			assert.Errorf(err, "getApplication with id %v (%t) returned no error altough it should", id, id)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "getApplication id was set to %v (%T) and should result in status code %d but code is %d", id, id, req.ShouldStatus, w.Code)

	}
}

func TestApi_getUser(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	_, err := mockups.AddUsersToDb(TestDatabase, TestUsers)
	assert.NoErrorf(err, "Adding users to database failed: %v", err)

	// No testing of invalid ids as that is tested in TestApi_getID already
	testCases := make(map[uint]tests.Request)
	testCases[500] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 404}
	testCases[1] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[2] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("id", id)
		user, err := getUser(c, TestDatabase)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			assert.Equalf(user.ID, id, "getUser id was set to %d but resulting app id is %d", id, user.ID)
			assert.NoErrorf(err, "getUser with id %v (%t) returned an error altough it should not: %v", id, id, err)
		} else {
			assert.Errorf(err, "getUser with id %v (%t) returned no error altough it should", id, id)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "getUser id was set to %v (%T) and should result in status code %d but code is %d", id, id, req.ShouldStatus, w.Code)
	}
}
