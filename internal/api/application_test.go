package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/pushbits/server/tests/mockups"
	"github.com/stretchr/testify/assert"
)

var TestApplicationHandler *ApplicationHandler
var TestUser *model.User

// Collect all created applications to check & delete them later
var SuccessAplications []model.Application

func TestMain(m *testing.M) {
	// Get main config and adapt
	config, err := mockups.ReadConfig("../../config.yml", true)
	if err != nil {
		cleanUp()
		log.Println("Can not read config: ", err)
		os.Exit(1)
	}

	config.Database.Connection = "pushbits-test.db"
	config.Database.Dialect = "sqlite3"

	// Set up test environment
	appHandler, err := getApplicationHandler(&config.Matrix)
	if err != nil {
		cleanUp()
		log.Println("Can not set up application handler: ", err)
		os.Exit(1)
	}

	TestApplicationHandler = appHandler

	// Run for each user
	for _, user := range mockups.GetUsers(config) {
		SuccessAplications = []model.Application{}
		TestUser = user
		m.Run()
	}
	cleanUp()
}

func TestApi_RegisterApplicationWithoutUser(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	reqWoUser := tests.Request{Name: "Invalid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test1", "strict_compatibility": true}`, Headers: map[string]string{"Content-Type": "application/json"}}
	_, c, err := reqWoUser.GetRequest()
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Panicsf(func() { TestApplicationHandler.CreateApplication(c) }, "CreateApplication did not panic altough user is not in context")

}

func TestApi_RgisterApplication(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testCases := make(map[int]tests.Request)
	testCases[400] = tests.Request{Name: "Invalid Form Data", Method: "POST", Endpoint: "/application", Data: "k=1&v=abc"}
	testCases[400] = tests.Request{Name: "Invalid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test1", "strict_compatibility": "oh yes"}`, Headers: map[string]string{"Content-Type": "application/json"}}
	testCases[200] = tests.Request{Name: "Valid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test2", "strict_compatibility": true}`, Headers: map[string]string{"Content-Type": "application/json"}}

	for statusCode, req := range testCases {
		var application model.Application
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("user", TestUser)
		TestApplicationHandler.CreateApplication(c)

		// Parse body only for successful requests
		if statusCode >= 200 && statusCode < 300 {
			body, err := ioutil.ReadAll(w.Body)
			assert.NoErrorf(err, "Can not read request body")
			if err != nil {
				continue
			}
			err = json.Unmarshal(body, &application)
			assert.NoErrorf(err, "Can not unmarshal request body")
			if err != nil {
				continue
			}

			SuccessAplications = append(SuccessAplications, application)
		}

		assert.Equalf(w.Code, statusCode, "CreateApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, statusCode, w.Code)
	}
}

func TestApi_GetApplications(t *testing.T) {
	var applications []model.Application

	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testCases := make(map[int]tests.Request)
	testCases[200] = tests.Request{Name: "Valid Request", Method: "GET", Endpoint: "/application"}

	for statusCode, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("user", TestUser)
		TestApplicationHandler.GetApplications(c)

		// Parse body only for successful requests
		if statusCode >= 200 && statusCode < 300 {
			body, err := ioutil.ReadAll(w.Body)
			assert.NoErrorf(err, "Can not read request body")
			if err != nil {
				continue
			}
			err = json.Unmarshal(body, &applications)
			assert.NoErrorf(err, "Can not unmarshal request body")
			if err != nil {
				continue
			}

			assert.Truef(validateAllApplications(applications), "Did not find application created previously")
			assert.Equalf(len(applications), len(SuccessAplications), "Created %d application(s) but got %d back", len(SuccessAplications), len(applications))
		}

		assert.Equalf(w.Code, statusCode, "GetApplications (Test case: \"%s\") should return status code %v but is %v.", req.Name, statusCode, w.Code)
	}
}

func TestApi_GetApplicationsWithoutUser(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testCase := tests.Request{Name: "Valid Request", Method: "GET", Endpoint: "/application"}

	_, c, err := testCase.GetRequest()
	if err != nil {
		t.Fatalf(err.Error())
	}

	assert.Panicsf(func() { TestApplicationHandler.GetApplications(c) }, "GetApplications did not panic altough user is not in context")

}

// GetApplicationHandler creates and returns an application handler
func getApplicationHandler(c *configuration.Matrix) (*ApplicationHandler, error) {
	db, err := mockups.GetEmptyDatabase()
	if err != nil {
		return nil, err
	}

	dispatcher, err := mockups.GetMatrixDispatcher(c.Homeserver, c.Username, c.Password)
	if err != nil {
		return nil, err
	}
	return &ApplicationHandler{
		DB: db,
		DP: dispatcher,
	}, nil
}

// True if all created applications are in list
func validateAllApplications(apps []model.Application) bool {
	for _, successApp := range SuccessAplications {
		foundApp := false
		for _, app := range apps {
			if app.ID == successApp.ID {
				foundApp = true
				break
			}
		}

		if !foundApp {
			return false
		}
	}

	return true
}

func cleanUp() {
	os.Remove("pushbits-test.db")
}
