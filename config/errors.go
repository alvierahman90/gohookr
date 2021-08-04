package config

import "fmt"

type requiredFieldError struct {
	fieldName   string
	serviceName string
}

func (e requiredFieldError) Error() string {
	return fmt.Sprintf("%v cannot be empty (%v)", e.fieldName, e.serviceName)
}
