package api

import (
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pushbits/server/internal/authentication/credentials"
	"github.com/pushbits/server/internal/configuration"
	"github.com/pushbits/server/internal/database"
	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/model"
	"github.com/pushbits/server/tests/mockups"
)

// TestContext holds all test-related objects
type TestContext struct {
	ApplicationHandler  *ApplicationHandler
	Users               []*model.User
	Database            *database.Database
	NotificationHandler *NotificationHandler
	UserHandler         *UserHandler
	Config              *configuration.Configuration
}

var GlobalTestContext *TestContext

func cleanup() {
	err := os.Remove("pushbits-test.db")
	if err != nil {
		log.L.Warnln("Cannot delete test database: ", err)
	}
}

func TestMain(m *testing.M) {
	cleanup()

	gin.SetMode(gin.TestMode)

	GlobalTestContext = CreateTestContext(nil)

	m.Run()

	cleanup()
}

// GetTestContext initializes and verifies all required test components
func GetTestContext(_ *testing.T) *TestContext {
	if GlobalTestContext == nil {
		GlobalTestContext = CreateTestContext(nil)
	}

	return GlobalTestContext
}

// CreateTestContext initializes and verifies all required test components
func CreateTestContext(_ *testing.T) *TestContext {
	ctx := &TestContext{}

	config := configuration.Configuration{}
	config.Database.Connection = "pushbits-test.db"
	config.Database.Dialect = "sqlite3"
	config.Crypto.Argon2.Iterations = 4
	config.Crypto.Argon2.Parallelism = 4
	config.Crypto.Argon2.Memory = 131072
	config.Crypto.Argon2.SaltLength = 16
	config.Crypto.Argon2.KeyLength = 32
	config.Admin.Name = "user"
	config.Admin.Password = "pushbits"
	ctx.Config = &config

	db, err := mockups.GetEmptyDatabase(ctx.Config.Crypto)
	if err != nil {
		cleanup()
		panic(fmt.Errorf("cannot set up database: %w", err))
	}
	ctx.Database = db

	ctx.ApplicationHandler = &ApplicationHandler{
		DB: ctx.Database,
		DP: &mockups.MockDispatcher{},
	}

	ctx.Users = mockups.GetUsers(ctx.Config)

	ctx.NotificationHandler = &NotificationHandler{
		DB: ctx.Database,
		DP: &mockups.MockDispatcher{},
	}

	ctx.UserHandler = &UserHandler{
		AH: ctx.ApplicationHandler,
		CM: credentials.CreateManager(false, ctx.Config.Crypto),
		DB: ctx.Database,
		DP: &mockups.MockDispatcher{},
	}

	return ctx
}
