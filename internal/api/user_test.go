package api

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
	}
}

func TestApi_GetUsers(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

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
	require.NoErrorf(err, "Failed to parse response body")
	err = json.Unmarshal(usersRaw, &users)
	require.NoErrorf(err, "Can not unmarshal users")

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

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
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

func TestApi_GetUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testCases := make(map[interface{}]tests.Request)
	testCases["abcde"] = tests.Request{Name: "Invalid id - string", Method: "GET", Endpoint: "/user/abcde", ShouldStatus: 500}
	testCases[uint(9999999)] = tests.Request{Name: "Unknown id", Method: "GET", Endpoint: "/user/99999999", ShouldStatus: 404}

	// Check if we can get all existing users
	for _, user := range TestUsers {
		testCases[user.ID] = tests.Request{
			Name:         "Valid user " + user.Name,
			Method:       "GET",
			Endpoint:     "/user/" + strconv.Itoa(int(user.ID)),
			ShouldStatus: 200,
			ShouldReturn: user,
		}
	}

	for id, testCase := range testCases {
		w, c, err := testCase.GetRequest()
		require.NoErrorf(err, "(Test case %s) Could not make request", testCase.Name)

		c.Set("id", id)
		TestUserHandler.GetUser(c)

		assert.Equalf(testCase.ShouldStatus, w.Code, "(Test case %s) Expected status code %d but have %d", testCase.Name, testCase.ShouldStatus, w.Code)

		// Check content for successful requests
		if testCase.ShouldReturn == 200 {
			user := &model.ExternalUser{}
			userBytes, err := ioutil.ReadAll(w.Body)
			require.NoErrorf(err, "(Test case %s) Can not read body", testCase.Name)
			err = json.Unmarshal(userBytes, user)
			require.NoErrorf(err, "(Test case %s) Can not unmarshal body", testCase.Name)

			shouldUser, ok := testCase.ShouldReturn.(*model.User)
			assert.Truef(ok, "(Test case %s) Successful response but no should response", testCase.Name)

			// Check if the returned user match
			assert.Equalf(user.ID, shouldUser.ID, "(Test case %s) User ID should be %d but is %d", testCase.Name, user.ID, shouldUser.ID)
			assert.Equalf(user.Name, shouldUser.Name, "(Test case %s) User name should be %s but is %s", testCase.Name, user.Name, shouldUser.Name)
			assert.Equalf(user.MatrixID, shouldUser.MatrixID, "(Test case %s) User matrix ID should be %s but is %s", testCase.Name, user.MatrixID, shouldUser.MatrixID)
			assert.Equalf(user.IsAdmin, shouldUser.IsAdmin, "(Test case %s) User is admin should be %v but is %v", testCase.Name, user.IsAdmin, shouldUser.IsAdmin)
		}
	}
}

func TestApi_DeleteUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	testCases := make(map[interface{}]tests.Request)
	testCases["abcde"] = tests.Request{Name: "Invalid user - string", Method: "DELETE", Endpoint: "/user/abcde", ShouldStatus: 500}
	testCases[uint(999999)] = tests.Request{Name: "Unknown user", Method: "DELETE", Endpoint: "/user/999999", ShouldStatus: 404}

	for _, user := range TestUsers {
		shouldStatus := 200
		testCases[user.ID] = tests.Request{
			Name:         "Valid user " + user.Name,
			Method:       "DELETE",
			Endpoint:     "/user/" + strconv.Itoa(int(user.ID)),
			ShouldStatus: shouldStatus,
		}
	}

	for id, testCase := range testCases {
		w, c, err := testCase.GetRequest()
		require.NoErrorf(err, "(Test case %s) Could not make request", testCase.Name)

		c.Set("id", id)
		TestUserHandler.DeleteUser(c)

		assert.Equalf(testCase.ShouldStatus, w.Code, "(Test case %s) Expected status code %d but have %d", testCase.Name, testCase.ShouldStatus, w.Code)
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
