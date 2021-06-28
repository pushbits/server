package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/pushbits/server/tests/mockups"
	"github.com/stretchr/testify/assert"
)

var TestApplicationHandler *ApplicationHandler
var TestUsers []*model.User
var TestDatabase *database.Database

// Collect all created applications to check & delete them later
var SuccessAplications map[uint][]model.Application

func TestMain(m *testing.M) {
	// Get main config and adapt
	config := &configuration.Configuration{}

	config.Database.Connection = "pushbits-test.db"
	config.Database.Dialect = "sqlite3"
	config.Crypto.Argon2.Iterations = 4
	config.Crypto.Argon2.Parallelism = 4
	config.Crypto.Argon2.Memory = 131072
	config.Crypto.Argon2.SaltLength = 16
	config.Crypto.Argon2.KeyLength = 32

	// Set up test environment
	db, err := mockups.GetEmptyDatabase(config.Crypto)
	if err != nil {
		cleanUp()
		log.Println("Can not set up database: ", err)
		os.Exit(1)
	}
	TestDatabase = db

	appHandler, err := getApplicationHandler(config)
	if err != nil {
		cleanUp()
		log.Println("Can not set up application handler: ", err)
		os.Exit(1)
	}

	TestApplicationHandler = appHandler
	TestUsers = mockups.GetUsers(config)
	SuccessAplications = make(map[uint][]model.Application)

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

func TestApi_RegisterApplication(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "Invalid Form Data", Method: "POST", Endpoint: "/application", Data: "k=1&v=abc", ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Invalid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test1", "strict_compatibility": "oh yes"}`, Headers: map[string]string{"Content-Type": "application/json"}, ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Valid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test2", "strict_compatibility": true}`, Headers: map[string]string{"Content-Type": "application/json"}, ShouldStatus: 200})

	for _, user := range TestUsers {
		SuccessAplications[user.ID] = make([]model.Application, 0)
		for _, req := range testCases {
			var application model.Application
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			TestApplicationHandler.CreateApplication(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
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

				SuccessAplications[user.ID] = append(SuccessAplications[user.ID], application)
			}

			assert.Equalf(w.Code, req.ShouldStatus, "CreateApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_GetApplications(t *testing.T) {
	var applications []model.Application

	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "Valid Request", Method: "GET", Endpoint: "/application", ShouldStatus: 200})

	for _, user := range TestUsers {
		for _, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			TestApplicationHandler.GetApplications(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
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

				assert.Truef(validateAllApplications(user, applications), "Did not find application created previously")
				assert.Equalf(len(applications), len(SuccessAplications[user.ID]), "Created %d application(s) but got %d back", len(SuccessAplications[user.ID]), len(applications))
			}

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplications (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
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

func TestApi_GetApplicationErrors(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	// Arbitrary test cases
	testCases := make(map[uint]tests.Request)
	testCases[0] = tests.Request{Name: "Requesting unknown application 0", Method: "GET", Endpoint: "/application/0", ShouldStatus: 404}
	testCases[5555] = tests.Request{Name: "Requesting unknown application 5555", Method: "GET", Endpoint: "/application/5555", ShouldStatus: 404}
	testCases[99999999999999999] = tests.Request{Name: "Requesting unknown application 99999999999999999", Method: "GET", Endpoint: "/application/99999999999999999", ShouldStatus: 404}

	for _, user := range TestUsers {
		for id, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			TestApplicationHandler.GetApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_GetApplication(t *testing.T) {
	var application model.Application

	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	// Previously generated applications
	for _, user := range TestUsers {
		for _, app := range SuccessAplications[user.ID] {
			req := tests.Request{Name: fmt.Sprintf("Requesting application %s (%d)", app.Name, app.ID), Method: "GET", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200}

			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			c.Set("id", app.ID)
			TestApplicationHandler.GetApplication(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
				body, err := ioutil.ReadAll(w.Body)
				assert.NoErrorf(err, "Can not read request body")
				if err != nil {
					continue
				}
				err = json.Unmarshal(body, &application)
				assert.NoErrorf(err, "Can not unmarshal request body: %v", err)
				if err != nil {
					continue
				}

				assert.Equalf(application.ID, app.ID, "Application ID should be %d but is %d", app.ID, application.ID)
				assert.Equalf(application.Name, app.Name, "Application Name should be %s but is %s", app.Name, application.Name)
				assert.Equalf(application.UserID, app.UserID, "Application user ID should be %d but is %d", app.UserID, application.UserID)

			}

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_UpdateApplication(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	for _, user := range TestUsers {
		testCases := make(map[uint]tests.Request)
		// Previously generated applications
		for _, app := range SuccessAplications[user.ID] {
			newName := app.Name + "-new_name"
			updateApp := model.UpdateApplication{
				Name: &newName,
			}
			updateAppBytes, err := json.Marshal(updateApp)
			assert.NoErrorf(err, "Error on marshaling updateApplication struct")

			// Valid
			testCases[app.ID] = tests.Request{Name: fmt.Sprintf("Update application (valid) %s (%d)", app.Name, app.ID), Method: "PUT", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200, Data: string(updateAppBytes), Headers: map[string]string{"Content-Type": "application/json"}}
			// Invalid
			testCases[app.ID] = tests.Request{Name: fmt.Sprintf("Update application (invalid) %s (%d)", app.Name, app.ID), Method: "PUT", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200, Data: "{}", Headers: map[string]string{"Content-Type": "application/json"}}
		}
		// Arbitrary test cases
		testCases[5555] = tests.Request{Name: "Update application 5555", Method: "PUT", Endpoint: "/application/5555", ShouldStatus: 404, Data: "random data"}
		testCases[5556] = tests.Request{Name: "Update application 5556", Method: "PUT", Endpoint: "/application/5556", ShouldStatus: 404, Data: `{"new_name": "new name"}`, Headers: map[string]string{"Content-Type": "application/json"}}

		for id, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			TestApplicationHandler.UpdateApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "UpdateApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_DeleteApplication(t *testing.T) {
	assert := assert.New(t)
	gin.SetMode(gin.TestMode)

	for _, user := range TestUsers {
		testCases := make(map[uint]tests.Request)
		// Previously generated applications
		for _, app := range SuccessAplications[user.ID] {
			testCases[app.ID] = tests.Request{Name: fmt.Sprintf("Delete application %s (%d)", app.Name, app.ID), Method: "DELETE", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200}
		}
		// Arbitrary test cases
		testCases[5555] = tests.Request{Name: "Delete application 5555", Method: "DELETE", Endpoint: "/application/5555", ShouldStatus: 404}

		for id, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatalf(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			TestApplicationHandler.DeleteApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "DeleteApplication (Test case: \"%s\") should return status code %v but is %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

// GetApplicationHandler creates and returns an application handler
func getApplicationHandler(c *configuration.Configuration) (*ApplicationHandler, error) {
	dispatcher := &mockups.MockDispatcher{}

	return &ApplicationHandler{
		DB: TestDatabase,
		DP: dispatcher,
	}, nil
}

// True if all created applications are in list
func validateAllApplications(user *model.User, apps []model.Application) bool {
	if _, ok := SuccessAplications[user.ID]; !ok {
		return len(apps) == 0
	}

	for _, successApp := range SuccessAplications[user.ID] {
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
