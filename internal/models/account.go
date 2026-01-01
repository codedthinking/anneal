package models

// Account represents a Fastmail account configuration
type Account struct {
	Name    string `yaml:"name"`
	Email   string `yaml:"email"`
	Default bool   `yaml:"default,omitempty"`
}
