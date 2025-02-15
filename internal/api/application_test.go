package api

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Collect all created applications to check & delete them later
var SuccessApplications = make(map[uint][]model.Application)

func TestApi_RegisterApplicationWithoutUser(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)

	reqWoUser := tests.Request{Name: "Invalid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test1", "strict_compatibility": true}`, Headers: map[string]string{"Content-Type": "application/json"}}
	_, c, err := reqWoUser.GetRequest()
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Panicsf(func() { ctx.ApplicationHandler.CreateApplication(c) }, "CreateApplication did not panic although user is not in context")
}

func TestApi_RegisterApplication(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)
	require := require.New(t)

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "Invalid Form Data", Method: "POST", Endpoint: "/application", Data: "k=1&v=abc", ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Invalid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test1", "strict_compatibility": "oh yes"}`, Headers: map[string]string{"Content-Type": "application/json"}, ShouldStatus: 400})
	testCases = append(testCases, tests.Request{Name: "Valid JSON Data", Method: "POST", Endpoint: "/application", Data: `{"name": "test2", "strict_compatibility": true}`, Headers: map[string]string{"Content-Type": "application/json"}, ShouldStatus: 200})

	for _, user := range ctx.Users {
		SuccessApplications[user.ID] = make([]model.Application, 0)
		for _, req := range testCases {
			var application model.Application
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			ctx.ApplicationHandler.CreateApplication(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
				body, err := io.ReadAll(w.Body)
				require.NoErrorf(err, "Cannot read request body")
				err = json.Unmarshal(body, &application)
				require.NoErrorf(err, "Cannot unmarshal request body")

				SuccessApplications[user.ID] = append(SuccessApplications[user.ID], application)
			}

			assert.Equalf(w.Code, req.ShouldStatus, "CreateApplication (Test case: \"%s\") Expected status code %v but received %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_GetApplications(t *testing.T) {
	ctx := GetTestContext(t)

	var applications []model.Application

	assert := assert.New(t)
	require := require.New(t)

	testCases := make([]tests.Request, 0)
	testCases = append(testCases, tests.Request{Name: "Valid Request", Method: "GET", Endpoint: "/application", ShouldStatus: 200})

	for _, user := range ctx.Users {
		for _, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			ctx.ApplicationHandler.GetApplications(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
				body, err := io.ReadAll(w.Body)
				require.NoErrorf(err, "Cannot read request body")
				err = json.Unmarshal(body, &applications)
				require.NoErrorf(err, "Cannot unmarshal request body")
				if err != nil {
					continue
				}

				assert.Truef(validateAllApplications(user, applications), "Did not find application created previously")
				assert.Equalf(len(applications), len(SuccessApplications[user.ID]), "Created %d application(s) but got %d back", len(SuccessApplications[user.ID]), len(applications))
			}

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplications (Test case: \"%s\") Expected status code %v but received %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_GetApplicationsWithoutUser(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)

	testCase := tests.Request{Name: "Valid Request", Method: "GET", Endpoint: "/application"}

	_, c, err := testCase.GetRequest()
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Panicsf(func() { ctx.ApplicationHandler.GetApplications(c) }, "GetApplications did not panic although user is not in context")
}

func TestApi_GetApplicationErrors(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)

	// Arbitrary test cases
	testCases := make(map[uint]tests.Request)
	testCases[0] = tests.Request{Name: "Requesting unknown application 0", Method: "GET", Endpoint: "/application/0", ShouldStatus: 404}
	testCases[5555] = tests.Request{Name: "Requesting unknown application 5555", Method: "GET", Endpoint: "/application/5555", ShouldStatus: 404}
	testCases[99999999999999999] = tests.Request{Name: "Requesting unknown application 99999999999999999", Method: "GET", Endpoint: "/application/99999999999999999", ShouldStatus: 404}

	for _, user := range ctx.Users {
		for id, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			ctx.ApplicationHandler.GetApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplication (Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_GetApplication(t *testing.T) {
	ctx := GetTestContext(t)

	var application model.Application

	assert := assert.New(t)
	require := require.New(t)

	// Previously generated applications
	for _, user := range ctx.Users {
		for _, app := range SuccessApplications[user.ID] {
			req := tests.Request{Name: fmt.Sprintf("Requesting application %s (%d)", app.Name, app.ID), Method: "GET", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200}

			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			c.Set("id", app.ID)
			ctx.ApplicationHandler.GetApplication(c)

			// Parse body only for successful requests
			if req.ShouldStatus >= 200 && req.ShouldStatus < 300 {
				body, err := io.ReadAll(w.Body)
				require.NoErrorf(err, "Cannot read request body")
				err = json.Unmarshal(body, &application)
				require.NoErrorf(err, "Cannot unmarshal request body: %v", err)

				assert.Equalf(application.ID, app.ID, "Application ID should be %d but is %d", app.ID, application.ID)
				assert.Equalf(application.Name, app.Name, "Application Name should be %s but is %s", app.Name, application.Name)
				assert.Equalf(application.UserID, app.UserID, "Application user ID should be %d but is %d", app.UserID, application.UserID)

			}

			assert.Equalf(w.Code, req.ShouldStatus, "GetApplication (Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_UpdateApplication(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)
	require := require.New(t)

	for _, user := range ctx.Users {
		testCases := make(map[uint]tests.Request)
		// Previously generated applications
		for _, app := range SuccessApplications[user.ID] {
			newName := app.Name + "-new_name"
			updateApp := model.UpdateApplication{
				Name: &newName,
			}
			updateAppBytes, err := json.Marshal(updateApp)
			require.NoErrorf(err, "Error on marshaling updateApplication struct")

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
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			ctx.ApplicationHandler.UpdateApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "UpdateApplication (Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

func TestApi_DeleteApplication(t *testing.T) {
	ctx := GetTestContext(t)

	assert := assert.New(t)

	for _, user := range ctx.Users {
		testCases := make(map[uint]tests.Request)
		// Previously generated applications
		for _, app := range SuccessApplications[user.ID] {
			testCases[app.ID] = tests.Request{Name: fmt.Sprintf("Delete application %s (%d)", app.Name, app.ID), Method: "DELETE", Endpoint: fmt.Sprintf("/application/%d", app.ID), ShouldStatus: 200}
		}
		// Arbitrary test cases
		testCases[5555] = tests.Request{Name: "Delete application 5555", Method: "DELETE", Endpoint: "/application/5555", ShouldStatus: 404}

		for id, req := range testCases {
			w, c, err := req.GetRequest()
			if err != nil {
				t.Fatal(err.Error())
			}

			c.Set("user", user)
			c.Set("id", id)
			ctx.ApplicationHandler.DeleteApplication(c)

			assert.Equalf(w.Code, req.ShouldStatus, "DeleteApplication (Test case: \"%s\") Expected status code %v but have %v.", req.Name, req.ShouldStatus, w.Code)
		}
	}
}

// True if all created applications are in list
func validateAllApplications(user *model.User, apps []model.Application) bool {
	if _, ok := SuccessApplications[user.ID]; !ok {
		return len(apps) == 0
	}

	for _, successApp := range SuccessApplications[user.ID] {
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
