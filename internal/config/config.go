package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/the9x/anneal/internal/models"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	serviceName = "tuimail"
	configDir   = ".config/tuimail"
	configFile  = "config.yaml"
)

// Config represents the application configuration
type Config struct {
	Accounts    []models.Account `yaml:"accounts"`
	Theme       string           `yaml:"theme"`
	Editor      string           `yaml:"editor"`
	PreviewPane bool             `yaml:"preview_pane"`
	Threading   bool             `yaml:"threading"`
	PageSize    int              `yaml:"page_size"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Theme:       "dark",
		Editor:      os.Getenv("EDITOR"),
		PreviewPane: true,
		Threading:   true,
		PageSize:    50,
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, configFile), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// GetToken retrieves the API token for an account from the system keyring
func GetToken(email string) (string, error) {
	return keyring.Get(serviceName, email)
}

// SetToken stores the API token for an account in the system keyring
func SetToken(email, token string) error {
	return keyring.Set(serviceName, email, token)
}

// DeleteToken removes the API token for an account from the system keyring
func DeleteToken(email string) error {
	return keyring.Delete(serviceName, email)
}

// DefaultAccount returns the default account or the first one
func (c *Config) DefaultAccount() *models.Account {
	for i := range c.Accounts {
		if c.Accounts[i].Default {
			return &c.Accounts[i]
		}
	}
	if len(c.Accounts) > 0 {
		return &c.Accounts[0]
	}
	return nil
}

// AddAccount adds a new account to the configuration
func (c *Config) AddAccount(name, email string, isDefault bool) error {
	// Check for duplicate
	for _, acc := range c.Accounts {
		if acc.Email == email {
			return fmt.Errorf("account %s already exists", email)
		}
	}

	// If this is default, unset others
	if isDefault {
		for i := range c.Accounts {
			c.Accounts[i].Default = false
		}
	}

	c.Accounts = append(c.Accounts, models.Account{
		Name:    name,
		Email:   email,
		Default: isDefault,
	})

	return nil
}
