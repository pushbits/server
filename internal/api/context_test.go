package api

import (
	"testing"

	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/pushbits/server/tests/mockups"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi_getID(t *testing.T) {
	GetTestContext(t)

	assert := assert.New(t)
	require := require.New(t)
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
			t.Fatal(err.Error())
		}

		c.Set("id", id)
		idReturned, err := getID(c)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			require.NoErrorf(err, "getId with id %v (%t) returned an error although it should not: %v", id, id, err)

			idUint, ok := id.(uint)
			if ok {
				assert.Equalf(idReturned, idUint, "getApi id was set to %d but result is %d", idUint, idReturned)
			}
		} else {
			assert.Errorf(err, "getId with id %v (%t) returned no error although it should", id, id)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "getApi id was set to %v (%T) and should result in status code %d but code is %d", id, id, req.ShouldStatus, w.Code)
	}
}

func TestApi_getApplication(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)
	require := require.New(t)

	applications := mockups.GetAllApplications()

	err := mockups.AddApplicationsToDb(ctx.Database, applications)
	if err != nil {
		log.L.Fatalln("Cannot add mock applications to database: ", err)
	}

	// No testing of invalid ids as that is tested in TestApi_getID already
	testCases := make(map[uint]tests.Request)
	testCases[500] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 404}
	testCases[1] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[2] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatal(err.Error())
		}

		c.Set("id", id)
		app, err := getApplication(c, ctx.Database)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			require.NoErrorf(err, "getApplication with id %v returned an unexpected error: %v", id, err)
			require.NotNilf(app, "Expected a valid app for id %v, but got nil", id)
			assert.Equalf(app.ID, id, "Expected app ID %d, but got %d", id, app.ID)
		} else {
			require.Errorf(err, "Expected an error for id %v, but got none", id)
			assert.Nilf(app, "Expected app to be nil for id %v, but got %+v", id, app)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "Expected status code %d for id %v, but got %d", req.ShouldStatus, id, w.Code)
	}
}

func TestApi_getUser(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)
	require := require.New(t)

	_, err := mockups.AddUsersToDb(ctx.Database, ctx.Users)
	assert.NoErrorf(err, "Adding users to database failed: %v", err)

	// No testing of invalid ids as that is tested in TestApi_getID already
	testCases := make(map[uint]tests.Request)
	testCases[500] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 404}
	testCases[1] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}
	testCases[2] = tests.Request{Name: "-", Method: "GET", Endpoint: "/", Data: "", ShouldStatus: 200}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatal(err.Error())
		}

		c.Set("id", id)
		user, err := getUser(c, ctx.Database)

		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			require.NoErrorf(err, "getUser with id %v returned an unexpected error: %v", id, err)
			require.NotNilf(user, "Expected a valid user for id %v, but got nil", id)
			assert.Equalf(user.ID, id, "Expected user ID %d, but got %d", id, user.ID)
		} else {
			require.Errorf(err, "Expected an error for id %v, but got none", id)
			assert.Nilf(user, "Expected user to be nil for id %v, but got %+v", id, user)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "Expected status code %d for id %v, but got %d", req.ShouldStatus, id, w.Code)

	}
}
