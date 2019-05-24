package config

import (
	"io/ioutil"
)

func LoadConfigFile(config_file_name string) ([]byte, error) {
	data, err := ioutil.ReadFile(config_file_name)
	if err != nil {
		return nil, err
	}
	return data, nil
}

