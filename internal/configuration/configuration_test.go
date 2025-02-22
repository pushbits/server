package configuration

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jinzhu/configor"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/pushbits/server/internal/log"
	"github.com/pushbits/server/internal/pberrors"
)

type Pair struct {
	Is     interface{}
	Should interface{}
}

func TestMain(m *testing.M) {
	testMode = true
	m.Run()
	cleanUp()
	os.Exit(0)
}

func TestConfiguration_GetMinimal(t *testing.T) {
	err := writeMinimalConfig()
	if err != nil {
		log.L.Fatalln("Cannot write minimal config file: ", err)
	}

	validateConfig(t)
}

func TestConfiguration_GetValid(t *testing.T) {
	assert := assert.New(t)

	err := writeValidConfig()
	if err != nil {
		log.L.Fatalln("Cannot write valid config file: ", err)
	}

	validateConfig(t)

	config := Get()

	expectedValues := make(map[string]Pair)
	expectedValues["config.Admin.MatrixID"] = Pair{config.Admin.MatrixID, "000000"}
	expectedValues["config.Matrix.Username"] = Pair{config.Matrix.Username, "default-username"}
	expectedValues["config.Matrix.Password"] = Pair{config.Matrix.Password, "default-password"}

	for name, pair := range expectedValues {
		assert.Equalf(pair.Is, pair.Should, fmt.Sprintf("%s should be %v but is %v", name, pair.Should, pair.Is))
	}
}

func TestConfiguration_GetEmpty(t *testing.T) {
	err := writeEmptyConfig()
	if err != nil {
		fmt.Println("Could not write empty config: ", err)
		os.Exit(1)
	}

	assert.Panicsf(t, func() { Get() }, "Get() did not panic although config is empty")
}

func TestConfiguration_GetInvalid(t *testing.T) {
	err := writeInvalidConfig()
	if err != nil {
		fmt.Println("Could not write invalid config: ", err)
		os.Exit(1)
	}

	assert.Panicsf(t, func() { Get() }, "Get() did not panic although config is empty")
}

func TestConfiguaration_ConfigFiles(t *testing.T) {
	files := configFiles()

	assert.Greater(t, len(files), 0)
	for _, file := range files {
		assert.Truef(t, strings.HasSuffix(file, ".yml"), "%s is no yaml file", file)
	}
}

// Checks if the values in the configuration are plausible
func validateConfig(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanicsf(func() { Get() }, "Get configuration should not panic")

	config := Get()
	asGreater := make(map[string]Pair)
	asGreater["config.Crypto.Argon2.Memory"] = Pair{config.Crypto.Argon2.Memory, uint32(0)}
	asGreater["config.Crypto.Argon2.Iterations"] = Pair{config.Crypto.Argon2.Iterations, uint32(0)}
	asGreater["config.Crypto.Argon2.SaltLength"] = Pair{config.Crypto.Argon2.SaltLength, uint32(0)}
	asGreater["config.Crypto.Argon2.KeyLength"] = Pair{config.Crypto.Argon2.KeyLength, uint32(0)}
	asGreater["config.Crypto.Argon2.Parallelism"] = Pair{config.Crypto.Argon2.Parallelism, uint8(0)}
	asGreater["config.HTTP.Port"] = Pair{config.HTTP.Port, 0}
	for name, pair := range asGreater {
		assert.Greaterf(pair.Is, pair.Should, fmt.Sprintf("%s should be > %v but is %v", name, pair.Should, pair.Is))
	}

	asFalse := make(map[string]bool)
	asFalse["config.Formatting.ColoredTitle"] = config.Formatting.ColoredTitle
	asFalse["config.Debug"] = config.Debug
	asFalse["config.Security.CheckHIBP"] = config.Security.CheckHIBP
	for name, value := range asFalse {
		assert.Falsef(value, fmt.Sprintf("%s should be false but is %t", name, value))
	}
}

type MinimalConfiguration struct {
	Admin struct {
		MatrixID string
	}
	Matrix struct {
		Username string
		Password string
	}
}

type InvalidConfiguration struct {
	Debug int
	HTTP  struct {
		ListenAddress bool
	}
	Admin struct {
		Name int
	}
	Formatting string
}

// Writes a minimal config to config.yml
func writeMinimalConfig() error {
	cleanUp()

	config := MinimalConfiguration{}
	config.Admin.MatrixID = "000000"
	config.Matrix.Username = "default-username"
	config.Matrix.Password = "default-password"

	configString, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile("config_unittest.yml", configString, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// Writes a config with default values to config.yml
func writeValidConfig() error {
	cleanUp()

	err := writeMinimalConfig()
	if err != nil {
		return err
	}

	// Load minimal config to get default values
	config := &Configuration{}
	err = configor.New(&configor.Config{
		Environment:          "production",
		ENVPrefix:            "PUSHBITS",
		ErrorOnUnmatchedKeys: true,
	}).Load(config, "config_unittest.yml")
	if err != nil {
		return err
	}

	config.Admin.MatrixID = "000000"
	config.Matrix.Username = "default-username"
	config.Matrix.Password = "default-password"

	configString, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile("config_unittest.yml", configString, 0o644)
	if err != nil {
		return err
	}

	return nil
}

// Writes a config that is empty
func writeEmptyConfig() error {
	cleanUp()

	err := os.WriteFile("config_unittest.yml", []byte(""), 0o644)
	if err != nil {
		return err
	}

	return nil
}

// Writes a config with invalid entries
func writeInvalidConfig() error {
	cleanUp()

	config := InvalidConfiguration{}
	config.Debug = 1337
	config.HTTP.ListenAddress = true
	config.Admin.Name = 23
	config.Formatting = "Nice"

	configString, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	err = os.WriteFile("config_unittest.yml", configString, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func cleanUp() {
	err := os.Remove("config_unittest.yml")
	if err != nil {
		log.L.Warnln("Cannot remove config file: ", err)
	}
}

func TestConfigurationValidation_ConfigTLSFilesInconsistent(t *testing.T) {
	assert := assert.New(t)

	c := Configuration{}
	c.Admin.MatrixID = "000000"
	c.Matrix.Username = "default-username"
	c.Matrix.Password = "default-password"
	c.HTTP.CertFile = "populated"
	c.HTTP.KeyFile = ""

	is := validateConfiguration(&c)
	should := pberrors.ErrConfigTLSFilesInconsistent
	assert.Equal(is, should, "validateConfiguration() should return ConfigTLSFilesInconsistent")
}
