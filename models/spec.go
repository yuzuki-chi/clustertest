package models

import (
	"encoding/json"
	"fmt"
)

type SpecType string

// Spec represents an infrastructure specification of clustered system.
// The implementations of Spec interface includes Provisioner specific data.
type Spec interface {
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
}

// InfraConfig represents current infrastructure configuration.
type InfraConfig interface {
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
}
