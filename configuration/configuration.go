package configuration

import (
	"github.com/jinzhu/configor"
)

// Configuration holds values that can be configured by the user.
type Configuration struct {
	Database struct {
		Dialect    string `default:"sqlite3"`
		Connection string `default:"pushbits.db"`
	}
	Admin struct {
		Name     string `default:"admin"`
		Password string `default:"admin"`
		MatrixID string `required:"true"`
	}
	Matrix struct {
		Homeserver string `default:"https://matrix.org"`
		Username   string `required:"true"`
		Password   string `required:"true"`
	}
}

func configFiles() []string {
	return []string{"config.yml"}
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

	return config
}
