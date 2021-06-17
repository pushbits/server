package mockups

import "github.com/pushbits/server/internal/model"

// GetApplication1 returns an application with id 1
func GetApplication1() *model.Application {
	return &model.Application{
		ID:     1,
		Token:  "1234567890abcdefghijklmn",
		UserID: 1,
		Name:   "App1",
	}
}

// GetApplication2 returns an application with id 2
func GetApplication2() *model.Application {
	return &model.Application{
		ID:     2,
		Token:  "0987654321xyzabcdefghij",
		UserID: 1,
		Name:   "App2",
	}
}

// GetAllApplications returns all mock-applications as a list
func GetAllApplications() []*model.Application {
	applications := make([]*model.Application, 0)
	applications = append(applications, GetApplication1())
	applications = append(applications, GetApplication2())

	return applications
}
