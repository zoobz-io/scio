package scio

// registrationConfig holds configuration applied during registration.
type registrationConfig struct {
	metadata Metadata
}

// RegistrationOption configures resource registration.
type RegistrationOption func(*registrationConfig)

// WithDescription sets the resource description.
func WithDescription(desc string) RegistrationOption {
	return func(c *registrationConfig) {
		c.metadata.Description = desc
	}
}

// WithVersion sets the resource version.
func WithVersion(ver string) RegistrationOption {
	return func(c *registrationConfig) {
		c.metadata.Version = ver
	}
}

// WithTag adds a tag to the resource metadata.
func WithTag(key, value string) RegistrationOption {
	return func(c *registrationConfig) {
		if c.metadata.Tags == nil {
			c.metadata.Tags = make(map[string]string)
		}
		c.metadata.Tags[key] = value
	}
}

// applyOptions applies registration options to a config.
func applyOptions(opts []RegistrationOption) registrationConfig {
	cfg := registrationConfig{
		metadata: Metadata{
			Tags: make(map[string]string),
		},
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
