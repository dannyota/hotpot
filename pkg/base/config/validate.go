package config

import "fmt"

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	// Temporal HostPort is required
	if c.Temporal.HostPort == "" {
		return fmt.Errorf("temporal.host_port is required")
	}

	// Database fields are required
	if c.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if c.Database.Port == 0 {
		return fmt.Errorf("database.port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database.user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database.dbname is required")
	}

	return nil
}
