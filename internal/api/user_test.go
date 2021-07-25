package api

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
)

func TestApi_CreateUser(t *testing.T) {
	assert := assert.New(t)

	testCases := make([]tests.Request, 0)

	// Add all test users
	for _, user := range TestUsers {
		createUser := &model.CreateUser{}
		createUser.ExternalUser.Name = user.Name
		createUser.ExternalUser.MatrixID = "@" + user.Name + ":matrix.org"
		createUser.ExternalUser.IsAdmin = user.IsAdmin
		createUser.UserCredentials.Password = TestConfig.Admin.Password

		testCase := tests.Request{
			Name:         "Already existing user " + user.Name,
			Method:       "POST",
			Endpoint:     "/user",
			Data:         createUser,
			Headers:      map[string]string{"Content-Type": "application/json"},
			ShouldStatus: 400,
		}
		testCases = append(testCases, testCase)

	}

	testCases = append(testCases, tests.Request{Name: "No data", Method: "POST", Endpoint: "/user", Data: "", ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Missing data urlencoded", Method: "POST", Endpoint: "/user", Data: "name=superuser&id=1&lol=5", ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Valid user urlencoded", Method: "POST", Endpoint: "/user", Data: "name=testuser1&matrix_id=%40testuser1%3Amatrix.org&is_admin=false&password=abcde", Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, ShouldStatus: 200})
	testCases = append(testCases, tests.Request{Name: "Valid user JSON encoded", Method: "POST", Endpoint: "/user", Data: `{"name": "testuser2", "matrix_id": "@testuser2@matrix.org", "is_admin": true, "password": "cdefg"}`, Headers: map[string]string{"Content-Type": "application/json"}, ShouldStatus: 200})

	for _, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		TestUserHandler.CreateUser(c)

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
	}
}

func TestApi_GetUsers(t *testing.T) {
	assert := assert.New(t)

	request := tests.Request{
		Method:   "GET",
		Endpoint: "/user",
	}

	w, c, err := request.GetRequest()
	if err != nil {
		t.Fatalf((err.Error()))
	}

	TestUserHandler.GetUsers(c)
	assert.Equalf(w.Code, 200, "Response code should be 200 but is %d", w.Code)

	// Get users from body
	users := make([]model.ExternalUser, 0)
	usersRaw, err := ioutil.ReadAll(w.Body)
	assert.NoErrorf(err, "Failed to parse response body")
	err = json.Unmarshal(usersRaw, &users)
	assert.NoErrorf(err, "Can not unmarshal users")

	// Check existence of all known users
	for _, user := range TestUsers {
		found := false
		for _, userExt := range users {
			if user.ID == userExt.ID && user.Name == userExt.Name {
				found = true
				break
			}
		}
		assert.Truef(found, "Did not find user %s", user.Name)
	}
}

func TestApi_UpdateUser(t *testing.T) {
	assert := assert.New(t)
	admin := getAdmin()

	testCases := make(map[uint]tests.Request)

	// Add all test users
	for _, user := range TestUsers {
		updateUser := &model.UpdateUser{}
		user.Name += "+1"
		user.IsAdmin = !user.IsAdmin

		updateUser.Name = &user.Name
		updateUser.IsAdmin = &user.IsAdmin

		testCases[uint(1)] = tests.Request{
			Name:         "Already existing user " + user.Name,
			Method:       "POST",
			Endpoint:     "/user",
			Data:         updateUser,
			Headers:      map[string]string{"Content-Type": "application/json"},
			ShouldStatus: 200,
		}
	}

	// Valid updates
	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("id", id)
		c.Set("user", admin)
		TestUserHandler.UpdateUser(c)

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
	}

	// Invalid without user set
	for id, req := range testCases {
		_, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("id", id)
		assert.Panicsf(func() { TestUserHandler.UpdateUser(c) }, "User not set should panic but did not")
	}
}

func getAdmin() *model.User {
	for _, user := range TestUsers {
		if user.IsAdmin {
			return user
		}
	}
	return nil
}
