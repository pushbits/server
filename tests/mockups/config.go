package mockups

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/pushbits/server/internal/configuration"
)

// ReadConfig copies the given filename to the current folder and parses it as a config file. RemoveFile indicates whether to remove the copied file or not
func ReadConfig(filename string, removeFile bool) (config *configuration.Configuration, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			err = errors.New("paniced while reading config")
		}
	}()

	if filename == "" {
		return nil, errors.New("empty filename")
	}

	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile("config.yml", file, 0644)
	if err != nil {
		return nil, err
	}

	config = configuration.Get()

	if removeFile {
		os.Remove("config.yml")
	}

	return config, nil
}
