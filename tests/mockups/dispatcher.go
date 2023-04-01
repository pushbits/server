package mockups

import (
	"fmt"

	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/model"
)

// MockDispatcher is a dispatcher used for testing - it does not need any storage interface
type MockDispatcher struct{}

// RegisterApplication mocks a functions to create a channel for an application.
func (*MockDispatcher) RegisterApplication(id uint, name, _ string) (string, error) {
	return fmt.Sprintf("%d-%s", id, name), nil
}

// DeregisterApplication mocks a function to delete a channel for an application.
func (*MockDispatcher) DeregisterApplication(_ *model.Application, _ *model.User) error {
	return nil
}

// UpdateApplication mocks a function to update a channel for an application.
func (*MockDispatcher) UpdateApplication(_ *model.Application, _ *configuration.RepairBehavior) error {
	return nil
}

// SendNotification mocks a function to send a notification to a given user.
func (*MockDispatcher) SendNotification(_ *model.Application, _ *model.Notification) (id string, err error) {
	return randStr(15), nil
}

// DeleteNotification mocks a function to send a notification to a given user that another notification is deleted
func (*MockDispatcher) DeleteNotification(_ *model.Application, _ *model.DeleteNotification) error {
	return nil
}
