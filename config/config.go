package config

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
