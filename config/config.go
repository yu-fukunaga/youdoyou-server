package config

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration
type Config struct {
	Port               string `envconfig:"PORT" default:"8081"`
	FirestoreProjectID string `envconfig:"FIRESTORE_PROJECT_ID" default:"youdoyou-intelligence"`
	NotionToken        string `envconfig:"NOTION_TOKEN" required:"true"`
	GoogleGenaiApiKey  string `envconfig:"GOOGLE_GENAI_API_KEY" required:"true"`
}

var (
	configInstance *Config
	once           sync.Once
)

// LoadConfig loads environment variables and returns the Config instance
func LoadConfig() *Config {
	once.Do(func() {
		// Load .env file if it exists
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, relying on environment variables")
		} else {
			log.Println(".env file loaded")
		}

		var cfg Config
		err := envconfig.Process("", &cfg)
		if err != nil {
			log.Fatalf("Failed to load environment variables: %v", err)
		}
		configInstance = &cfg
	})
	return configInstance
}

// NewConfig compatibility alias (optional, but good for refactoring)
func NewConfig() *Config {
	return LoadConfig()
}
