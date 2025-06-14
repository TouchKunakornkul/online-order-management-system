package config

// Application configuration struct and loader.

type Config struct {
	PostgresDSN string
	// Add more config fields as needed
}

func LoadConfig() (*Config, error) {
	// Load config from env, file, etc.
	return &Config{}, nil
}
