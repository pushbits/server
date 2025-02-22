// Package configuration provides definitions and functionality related to the configuration.
package configuration

import (
	"github.com/jinzhu/configor"

	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/pberrors"
)

// testMode indicates if the package is run in test mode
var testMode bool

// Argon2Config holds the parameters used for creating hashes with Argon2.
type Argon2Config struct {
	Memory      uint32 `default:"131072"`
	Iterations  uint32 `default:"4"`
	Parallelism uint8  `default:"4"`
	SaltLength  uint32 `default:"16"`
	KeyLength   uint32 `default:"32"`
}

// CryptoConfig holds the parameters used for creating hashes.
type CryptoConfig struct {
	Argon2 Argon2Config
}

// Formatting holds additional parameters used for formatting messages
type Formatting struct {
	ColoredTitle bool `default:"false"`
}

// Matrix holds credentials for a matrix account
type Matrix struct {
	Homeserver string `default:"https://matrix.org"`
	Username   string `required:"true"`
	Password   string `required:"true"`
}

// Alertmanager holds information on how to parse alertmanager calls
type Alertmanager struct {
	AnnotationTitle   string `default:"title"`
	AnnotationMessage string `default:"message"`
}

// RepairBehavior holds information on how repair applications.
type RepairBehavior struct {
	ResetRoomName  bool `default:"true"`
	ResetRoomTopic bool `default:"true"`
}

// Configuration holds values that can be configured by the user.
type Configuration struct {
	Debug bool `default:"false"`
	HTTP  struct {
		ListenAddress  string   `default:""`
		Port           int      `default:"8080"`
		TrustedProxies []string `default:"[]"`
		CertFile       string   `default:""`
		KeyFile        string   `default:""`
	}
	Database struct {
		Dialect    string `default:"sqlite3"`
		Connection string `default:"pushbits.db"`
	}
	Admin struct {
		Name     string `default:"admin"`
		Password string `default:"admin"`
		MatrixID string `required:"true"`
	}
	Matrix   Matrix
	Security struct {
		CheckHIBP bool `default:"false"`
	}
	Crypto         CryptoConfig
	Formatting     Formatting
	Alertmanager   Alertmanager
	RepairBehavior RepairBehavior
}

func configFiles() []string {
	if testMode {
		return []string{"config_unittest.yml"}
	}
	return []string{"config.yml"}
}

func validateHTTPConfiguration(c *Configuration) error {
	certAndKeyEmpty := (c.HTTP.CertFile == "" && c.HTTP.KeyFile == "")
	certAndKeyPopulated := (c.HTTP.CertFile != "" && c.HTTP.KeyFile != "")

	if !certAndKeyEmpty && !certAndKeyPopulated {
		return pberrors.ErrConfigTLSFilesInconsistent
	}

	return nil
}

func validateConfiguration(c *Configuration) error {
	return validateHTTPConfiguration(c)
}

// Get returns the configuration extracted from env variables or config file.
func Get() *Configuration {
	config := &Configuration{}

	err := configor.New(&configor.Config{
		Environment:          "production",
		ENVPrefix:            "PUSHBITS",
		ErrorOnUnmatchedKeys: true,
	}).Load(config, configFiles()...)
	if err != nil {
		panic(err)
	}

	if err := validateConfiguration(config); err != nil {
		log.L.Fatal(err)
	}

	return config
}
