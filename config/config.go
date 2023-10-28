package config

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// The struct that represents the config.json file
type Config struct {
	ListenAddress string
	Services      map[string]struct {
		Script                       Command
		Secret                       string
		SignaturePrefix              string
		SignatureHeader              string
		DisableSignatureVerification bool
		Tests                        []Command
	}
}

// Check that all required fields are filled in
func (c Config) Validate() error {
	if c.ListenAddress == "" {
		return requiredFieldError{"ListenAddress", ""}
	}

	for serviceName, service := range c.Services {
		if service.Script.Program == "" {
			return requiredFieldError{"Script.Program", serviceName}
		}
		if !service.DisableSignatureVerification && service.SignatureHeader == "" {
			return requiredFieldError{"SignatureHeader", serviceName}
		}
		if !service.DisableSignatureVerification && service.Secret == "" {
			return requiredFieldError{"Secret", serviceName}
		}
	}

	return nil
}

func (c *Config) Load(config_filename string) error {

	raw_config, err := ioutil.ReadFile(config_filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw_config, &c)
	if err == nil {
		return c.Validate()
	}

	err = yaml.Unmarshal(raw_config, &c)
	if err == nil {
		return c.Validate()
	}

	return err
}
