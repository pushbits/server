package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi_SuccessOrAbort(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	testCases := make(map[error]tests.Request)
	testCases[errors.New("")] = tests.Request{Name: "Empty Error - 500", Endpoint: "/", ShouldStatus: 500}
	testCases[errors.New("this is an error")] = tests.Request{Name: "Error - 500", Endpoint: "/", ShouldStatus: 500}
	testCases[errors.New("this is an error")] = tests.Request{Name: "Error - 200", Endpoint: "/", ShouldStatus: 200}
	testCases[errors.New("this is an error")] = tests.Request{Name: "Error - 404", Endpoint: "/", ShouldStatus: 404}
	testCases[errors.New("this is an error")] = tests.Request{Name: "Error - 1001", Endpoint: "/", ShouldStatus: 1001}
	testCases[nil] = tests.Request{Name: "No Error - 1001", Endpoint: "/", ShouldStatus: 1001}
	testCases[nil] = tests.Request{Name: "No Error - 404", Endpoint: "/", ShouldStatus: 404}

	for forcedErr, testCase := range testCases {
		w, c, err := testCase.GetRequest()
		require.NoErrorf(err, "(Test case %s) Could not make request", testCase.Name)

		aborted := successOrAbort(c, testCase.ShouldStatus, forcedErr)

		if forcedErr != nil {
			assert.Equalf(testCase.ShouldStatus, w.Code, "(Test case %s) Expected status code %d but have %d", testCase.Name, testCase.ShouldStatus, w.Code)
		}

		assert.Equalf(forcedErr == nil, aborted, "(Test case %s) Expected %v but have %v", testCase.Name, forcedErr == nil, aborted)
	}
}

func TestApi_IsCurrentUser(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	for _, user := range TestUsers {
		testCases := make(map[uint]tests.Request)

		testCases[user.ID] = tests.Request{Name: fmt.Sprintf("User %s - success", user.Name), Endpoint: "/", ShouldStatus: 200}
		testCases[uint(49786749859)] = tests.Request{Name: fmt.Sprintf("User %s - fail", user.Name), Endpoint: "/", ShouldStatus: 403}

		for id, testCase := range testCases {
			w, c, err := testCase.GetRequest()
			require.NoErrorf(err, "(Test case %s) Could not make request", testCase.Name)

			c.Set("user", user)
			isCurrentUser := isCurrentUser(c, id)

			if testCase.ShouldStatus == 200 {
				assert.Truef(isCurrentUser, "(Test case %s) Should be true but is false", testCase.Name)
			} else {
				assert.Falsef(isCurrentUser, "(Test case %s) Should be false but is true", testCase.Name)
				assert.Equalf(w.Code, testCase.ShouldStatus, "(Test case %s) Expected status code %d but have %d", testCase.Name, testCase.ShouldStatus, w.Code)
			}
		}
	}

}
