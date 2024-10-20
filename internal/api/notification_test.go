package api

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApi_CreateNotification(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	gin.SetMode(gin.TestMode)

	testApplication := model.Application{
		ID:       1,
		Token:    "123456",
		UserID:   1,
		Name:     "Test Application",
		MatrixID: "@testuser:test.de",
	}

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "Valid with message", Method: "POST", Endpoint: "/message?token=123456&message=testmessage", ShouldStatus: 200, ShouldReturn: model.Notification{Message: "testmessage", Title: "Test Application"}})
	testCases = append(testCases, tests.Request{Name: "Valid with message and title", Method: "POST", Endpoint: "/message?token=123456&message=testmessage&title=abcdefghijklmnop", ShouldStatus: 200, ShouldReturn: model.Notification{Message: "testmessage", Title: "abcdefghijklmnop"}})
	testCases = append(testCases, tests.Request{Name: "Valid with message, title and priority", Method: "POST", Endpoint: "/message?token=123456&message=testmessage&title=abcdefghijklmnop&priority=3", ShouldStatus: 200, ShouldReturn: model.Notification{Message: "testmessage", Title: "abcdefghijklmnop", Priority: 3}})
	testCases = append(testCases, tests.Request{Name: "Invalid with wrong field message2", Method: "POST", Endpoint: "/message?token=123456&message2=testmessage", ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "No form data", Method: "POST", Endpoint: "/message", ShouldStatus: 400})

	for _, req := range testCases {
		var notification model.Notification
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatal(err.Error())
		}

		c.Set("app", &testApplication)
		TestNotificationHandler.CreateNotification(c)

		// Parse body only for successful requests
		if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
			body, err := io.ReadAll(w.Body)
			require.NoErrorf(err, "Can not read request body")
			err = json.Unmarshal(body, &notification)
			require.NoErrorf(err, "Can not unmarshal request body")

			shouldNotification, ok := req.ShouldReturn.(model.Notification)
			assert.Truef(ok, "(Test case %s) Type mismatch can not test further", req.Name)

			assert.Greaterf(len(notification.ID), 1, "(Test case %s) Notification id is not set correctly with \"%s\"", req.Name, notification.ID)

			assert.Equalf(shouldNotification.Message, notification.Message, "(Test case %s) Notification message should be %s but is %s", req.Name, shouldNotification.Message, notification.Message)
			assert.Equalf(shouldNotification.Title, notification.Title, "(Test case %s) Notification title should be %s but is %s", req.Name, shouldNotification.Title, notification.Title)
			assert.Equalf(shouldNotification.Priority, notification.Priority, "(Test case %s) Notification priority should be %s but is %s", req.Name, shouldNotification.Priority, notification.Priority)
		}

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
	}
}

func TestApi_DeleteNotification(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testApplication := model.Application{
		ID:       1,
		Token:    "123456",
		UserID:   1,
		Name:     "Test Application",
		MatrixID: "@testuser:test.de",
	}

	testCases := make(map[interface{}]tests.Request)
	testCases["1"] = tests.Request{Name: "Valid numeric string", Method: "DELETE", Endpoint: "/message?token=123456&message=testmessage", ShouldStatus: 200}
	testCases["abcde"] = tests.Request{Name: "Valid string", Method: "DELETE", Endpoint: "/message?token=123456&message=testmessage", ShouldStatus: 200}
	testCases[123456] = tests.Request{Name: "Invalid int", Method: "DELETE", Endpoint: "/message?token=123456&message=testmessage", ShouldStatus: 500}

	for id, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatal(err.Error())
		}

		c.Set("app", &testApplication)
		c.Set("messageid", id)
		TestNotificationHandler.DeleteNotification(c)

		assert.Equalf(w.Code, req.ShouldStatus, "(Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
	}
}
