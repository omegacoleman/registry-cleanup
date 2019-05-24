package config

import (
	"gopkg.in/yaml.v2"
)

type CleanupConfig struct {
	RegistryURL string `yaml:"registry_url"`
	RepositoryArray []string `yaml:"repos,flow"`
	MatchRules []string `yaml:"cleanup,flow"`
	ExceptRules []string `yaml:"except,flow"`
	Auth *struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth,omitempty"`
}

func Parse(in []byte) (*CleanupConfig, error) {
	config := CleanupConfig{}

	err := yaml.Unmarshal(in, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

