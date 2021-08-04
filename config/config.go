package config

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	ListenAddress string
	Services      map[string]struct {
		Script          Command
		Secret          string
		SignatureHeader string
		Tests           []Command
	}
}

func (c Config) Validate() error {
	if c.ListenAddress == "" {
		return requiredFieldError{"ListenAddress", ""}
	}

	jsonbytes, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println(string(jsonbytes))

	for serviceName, service := range c.Services {
		if service.Script.Program == "" {
			return requiredFieldError{"Script.Program", serviceName}
		}
		if service.SignatureHeader == "" {
			return requiredFieldError{"SignatureHeader", serviceName}
		}
		if service.Secret == "" {
			return requiredFieldError{"Secret", serviceName}
		}
	}

	return nil
}
