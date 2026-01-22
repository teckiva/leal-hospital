package medierror

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// ErrorRegistryKey is the key used to store the error registry in Gin context
const ErrorRegistryKey = "error_registry"

// ErrorConfig represents the structure of error configuration from YAML
type ErrorConfig struct {
	Message        string `yaml:"message"`
	DisplayMessage string `yaml:"display_message"`
	DeclineType    string `yaml:"decline_type"`
	Source         string `yaml:"source"`
}

// ErrorRegistry manages error configurations loaded from YAML
type ErrorRegistry struct {
	errors map[ErrorCode]*ErrorConfig
	mu     sync.RWMutex
}

// globalRegistry is the singleton instance for non-HTTP contexts
var (
	globalRegistry     *ErrorRegistry
	globalRegistryOnce sync.Once
	globalRegistryErr  error
)

// InitErrorRegistry initializes the error registry and registers it as Gin middleware
func InitErrorRegistry(router *gin.Engine) error {
	registry, err := loadErrorRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize error registry: %w", err)
	}

	// Set global registry
	globalRegistryOnce.Do(func() {
		globalRegistry = registry
		globalRegistryErr = nil
	})

	// Register middleware to inject registry into context
	router.Use(func(c *gin.Context) {
		c.Set(ErrorRegistryKey, registry)
		c.Next()
	})

	return nil
}

// loadErrorRegistry loads error configurations from YAML file
func loadErrorRegistry() (*ErrorRegistry, error) {
	// Get the config file path
	configPath := getConfigPath()

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read errors.yaml: %w", err)
	}

	// Parse YAML with numeric string keys
	var errorsMap map[string]*ErrorConfig
	if err := yaml.Unmarshal(data, &errorsMap); err != nil {
		return nil, fmt.Errorf("failed to parse errors.yaml: %w", err)
	}

	// Convert to ErrorCode keys
	errors := make(map[ErrorCode]*ErrorConfig)
	for code, config := range errorsMap {
		errors[ErrorCode(code)] = config
	}

	registry := &ErrorRegistry{
		errors: errors,
	}

	return registry, nil
}

// getConfigPath returns the path to the errors.yaml file
func getConfigPath() string {
	// Try to find config/errors.yaml relative to current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path
		return "config/errors.yaml"
	}

	// Check if we're in the server directory
	configPath := filepath.Join(cwd, "config", "errors.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Check if we're in a subdirectory and need to go up
	configPath = filepath.Join(cwd, "..", "config", "errors.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Check if we're in the root and need to go into server
	configPath = filepath.Join(cwd, "server", "config", "errors.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	// Default fallback
	return "config/errors.yaml"
}

// GetError retrieves an error configuration by code
func (r *ErrorRegistry) GetError(code ErrorCode) (*ErrorConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.errors[code]
	if !ok {
		return nil, fmt.Errorf("error code %s not found in registry", code)
	}

	return config, nil
}

// GetErrorWithDefault retrieves an error configuration by code with a default fallback
func (r *ErrorRegistry) GetErrorWithDefault(code ErrorCode, defaultMsg string) *ErrorConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.errors[code]
	if !ok {
		return &ErrorConfig{
			Message:        defaultMsg,
			DisplayMessage: defaultMsg,
			DeclineType:    "TD",
			Source:         "INTERNAL",
		}
	}

	return config
}

// HasError checks if an error code exists in the registry
func (r *ErrorRegistry) HasError(code ErrorCode) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.errors[code]
	return ok
}

// GetAllErrors returns all error configurations (useful for debugging)
func (r *ErrorRegistry) GetAllErrors() map[ErrorCode]*ErrorConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modifications
	errorsCopy := make(map[ErrorCode]*ErrorConfig, len(r.errors))
	maps.Copy(errorsCopy, r.errors)

	return errorsCopy
}

// Count returns the number of errors in the registry
func (r *ErrorRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.errors)
}

// GetErrorsBySource returns all errors for a specific source
func (r *ErrorRegistry) GetErrorsBySource(source string) map[ErrorCode]*ErrorConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[ErrorCode]*ErrorConfig)
	for code, config := range r.errors {
		if config.Source == source {
			result[code] = config
		}
	}

	return result
}

// GetErrorRegistry retrieves the error registry from Gin context
func GetErrorRegistry(c *gin.Context) *ErrorRegistry {
	if c == nil {
		return GetGlobalRegistry()
	}

	registry, exists := c.Get(ErrorRegistryKey)
	if !exists {
		return GetGlobalRegistry()
	}

	if reg, ok := registry.(*ErrorRegistry); ok {
		return reg
	}

	return GetGlobalRegistry()
}

// GetGlobalRegistry returns the global registry instance for non-HTTP contexts
func GetGlobalRegistry() *ErrorRegistry {
	return globalRegistry
}

// InitializeErrorRegistryStandalone initializes the error registry without Gin
// This is useful for testing or non-HTTP contexts
func InitializeErrorRegistryStandalone() (*ErrorRegistry, error) {
	var err error
	globalRegistryOnce.Do(func() {
		globalRegistry, err = loadErrorRegistry()
		globalRegistryErr = err
	})

	if err != nil {
		return nil, err
	}
	if globalRegistryErr != nil {
		return nil, globalRegistryErr
	}

	return globalRegistry, nil
}

// ResetGlobalRegistry resets the global registry (useful for testing)
func ResetGlobalRegistry() {
	globalRegistry = nil
	globalRegistryErr = nil
	globalRegistryOnce = sync.Once{}
}
