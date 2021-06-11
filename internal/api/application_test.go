package api

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/tests"
	"github.com/pushbits/server/tests/mockups"
	"github.com/stretchr/testify/assert"
)

var TestApplicationHandler *ApplicationHandler
var TestConfig *configuration.Configuration

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
	TestConfig = config

	// Set up test environment
	appHandler, err := getApplicationHandler(&TestConfig.Matrix)
	if err != nil {
		cleanUp()
		log.Println("Can not set up application handler: ", err)
		os.Exit(1)
	}

	TestApplicationHandler = appHandler

	// Run
	m.Run()
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

	user := mockups.GetAdminUser(TestConfig)

	for statusCode, req := range testCases {
		w, c, err := req.GetRequest()
		if err != nil {
			t.Fatalf(err.Error())
		}

		c.Set("user", user)

		TestApplicationHandler.CreateApplication(c)

		assert.Equalf(w.Code, statusCode, fmt.Sprintf("CreateApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, statusCode, w.Code))
	}
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

func cleanUp() {
	os.Remove("pushbits-test.db")
}
