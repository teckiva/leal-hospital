package config

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/krazybee/kb-pay/cloud/awsSecretManager"
	"github.com/krazybee/kb-pay/domain"
	"github.com/krazybee/kb-pay/logger"
	"github.com/spf13/viper"
)

// AppConfig defines application configuration
type AppConfig struct {
	Environment string
	Database    DBConfig
	YPDatabase  YPDBConfig
	Redis       RedisConfig
}

// DBConfig holds database configuration
type DBConfig struct {
	Type     string // "mysql", "memory", etc.
	Host     string
	Port     string
	Username string
	Password string
	Database string
	DSN      string
}

// YPDBConfig holds KB database configuration
type YPDBConfig struct {
	Type     string // "mysql", "memory", etc.
	Host     string
	Port     string
	Username string
	Password string
	Database string
	DSN      string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Type     string // "redis", "memory", etc.
	Host     string
	Port     string
	UserName string
	Password string
	Db       string
	URL      string
}

// LoadConfig initializes the configuration for the service.
func LoadConfig(env string) *AppConfig {
	// If env is empty, use environment variable or default
	if env == "" {
		env = viper.GetString("APP_ENV")
		if env == "" {
			env = "production"
		}
	}

	var configPath string
	flag.StringVar(&configPath, "config", "", "path to config file2")
	flag.Parse()

	viper.AutomaticEnv()
	// if this call is a lambda call, then get the data from AWS secrets manager and set the values in viper
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" && os.Getenv("SECRET_NAME") != "" {
		secretName := os.Getenv("SECRET_NAME")
		// call here cloud awsSecretManager.GetSecret(secretName)
		sm, err := awsSecretManager.NewSecretsManager()
		if err != nil {
			logger.Error("Secret Ininitialization Err", err)
			return &AppConfig{}
		}
		secretMap, err := sm.GetSecret(context.Background(), secretName)
		if err != nil {
			logger.Error("Secret Retrieval Err", err)
			return &AppConfig{}
		}
		for key, value := range secretMap {
			viper.Set(key, value)
		}

		return &AppConfig{
			Environment: env,
			Database: DBConfig{
				Type:     viper.GetString("DATABASE_TYPE"),
				Host:     viper.GetString("PAY_MYSQL_DB_HOST"),
				Port:     viper.GetString("PAY_MYSQL_DB_PORT"),
				Username: viper.GetString("PAY_MYSQL_DB_USERNAME"),
				Password: viper.GetString("PAY_MYSQL_DB_PASSWORD"),
				Database: viper.GetString("PAY_MYSQL_DB_SCHEMA"),
			},
			YPDatabase: YPDBConfig{
				Type:     viper.GetString("DATABASE_TYPE"),
				Host:     viper.GetString("YP_MYSQL_DB_HOST"),
				Port:     viper.GetString("YP_MYSQL_DB_PORT"),
				Username: viper.GetString("YP_MYSQL_DB_USERNAME"),
				Password: viper.GetString("YP_MYSQL_DB_PASSWORD"),
				Database: viper.GetString("YP_MYSQL_DB_SCHEMA"),
			},
			Redis: RedisConfig{
				Type:     viper.GetString("REDIS_TYPE"),
				Host:     viper.GetString("PAY_REDIS_HOST"),
				Port:     viper.GetString("PAY_REDIS_PORT"),
				UserName: viper.GetString("PAY_REDIS_USERNAME"),
				Password: viper.GetString("PAY_REDIS_PASSWORD"),
				Db:       viper.GetString("PAY_REDIS_DB"),
			},
		}

		// return viper.ReadInConfig()
	}

	// if this call is a local call, then get the data from the local config file and set the values in viper
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		configPath = os.Getenv("CONFIG_PATH")
		if configPath != "" {
			viper.AddConfigPath(configPath)
		}
		_, b, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(b)

		logger.Info("BasePath", basePath)
		viper.AddConfigPath(filepath.Join(basePath)) // Look for .env in project root/config
		viper.AddConfigPath(".")
		viper.SetConfigName(".env")
		viper.SetConfigType("env")
	}

	// Set defaults
	viper.SetDefault("DATABASE_TYPE", "mysql")
	viper.SetDefault("REDIS_TYPE", "redis")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found, use environment variables only
		logger.Error("Config file not found, using environment variables only: %v", err)
	}

	// Create config instance
	config := &AppConfig{
		Environment: env,
		Database: DBConfig{
			Type:     viper.GetString("DATABASE_TYPE"),
			Host:     viper.GetString("DATABASE_HOST"),
			Port:     viper.GetString("DATABASE_PORT"),
			Username: viper.GetString("DATABASE_USERNAME"),
			Password: viper.GetString("DATABASE_PASSWORD"),
			Database: viper.GetString("DATABASE_NAME"),
		},
		YPDatabase: YPDBConfig{
			Type:     viper.GetString("DATABASE_TYPE"),
			Host:     viper.GetString("YP_MYSQL_DB_HOST"),
			Port:     viper.GetString("YP_MYSQL_DB_PORT"),
			Username: viper.GetString("YP_MYSQL_DB_USERNAME"),
			Password: viper.GetString("YP_MYSQL_DB_PASSWORD"),
			Database: viper.GetString("YP_MYSQL_DB_SCHEMA"),
		},
		Redis: RedisConfig{
			Type:     viper.GetString("REDIS_TYPE"),
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			UserName: viper.GetString("PAY_REDIS_USERNAME"),
			Password: viper.GetString("PAY_REDIS_PASSWORD"),
			Db:       viper.GetString("PAY_REDIS_DB"),
		},
	}

	// Build connection strings
	if config.Database.Type == "mysql" {
		config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
		)
	}

	if config.YPDatabase.Type == "mysql" {
		config.YPDatabase.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.YPDatabase.Username,
			config.YPDatabase.Password,
			config.YPDatabase.Host,
			config.YPDatabase.Port,
			config.YPDatabase.Database,
		)
	}

	if config.Redis.Type == "redis" {
		config.Redis.URL = fmt.Sprintf("redis://%s:%s",
			config.Redis.Host,
			config.Redis.Port,
		)
	}

	return config
}

type PSPConfig struct {
	Merchants []MerchantConfig `yaml:"merchant"`
}

type MerchantConfig struct {
	ID         string `yaml:"id"`
	ChannelID  string `yaml:"channel_id"`
	KID        string `yaml:"kid"`
	PrivateKey string `yaml:"private_key"`
	PublicKey  string `yaml:"public_key"`
	BaseUrl    string `yaml:"base_url"`
}

// LoadJuspayAxisConfig initializes the configuration for juspay gateway.
func LoadJuspayAxisConfig() *domain.MerchantConfig {
	merchantConfigs := domain.MerchantConfig{
		MerchantID:                viper.GetString("JUSPAY_KREDITBEE_MERCHANT_ID"),
		ChannelID:                 viper.GetString("JUSPAY_KREDITBEE_MERCHANT_CHANNEL_ID"),
		Kid:                       viper.GetString("JUSPAY_KREDITBEE_MERCHANT_KID"),
		BaseUrl:                   viper.GetString("JUSPAY_AXIS_BASE_URL"),
		ProxyUrl:                  viper.GetString("SQUID_PROXY_URL"),
		PspEncryption:             viper.GetString("PSP_ENCRYPTION"),
		JweKid:                    viper.GetString("JUSPAY_JWE_KREDITBEE_MERCHANT_KID"),
		ValidateApiRespSignature:  viper.GetString("KB_PAY_VALIDATE_API_RESPONSE_SIGNATURE"),
		RawCallbackPayloadEnabled: viper.GetString("KB_PAY_RAW_CALLBACK_PAYLOAD_ENABLED"),
	}

	// Check if a path is provided for the private key and ues it if available
	// Helper function to read from file path or use direct value
	getValueFromFileOrDirect := func(pathKey, valueKey string) string {
		filePath := viper.GetString(pathKey)
		if filePath != "" {
			content, err := readFile(filePath)
			if err != nil {
				logger.Error("Error reading file %s: %v", filePath, err)
				return viper.GetString(valueKey)
			}
			return content
		}
		return viper.GetString(valueKey)
	}

	// Get all required keys using the helper function
	merchantConfigs.PrivateKey = getValueFromFileOrDirect("JUSPAY_KREDITBEE_PRIVATE_KEY_PATH", "JUSPAY_KREDITBEE_PRIVATE_KEY")
	merchantConfigs.JuspayAPIKey = getValueFromFileOrDirect("JUSPAY_API_KEY_PATH", "JUSPAY_API_KEY")
	merchantConfigs.JwePrivateKey = getValueFromFileOrDirect("JUSPAY_KREDITBEE_JWE_PRIVATE_KEY_PATH", "JUSPAY_KREDITBEE_JWE_PRIVATE_KEY")
	merchantConfigs.JwePublicKey = getValueFromFileOrDirect("JUSPAY_JWE_PUBLIC_KEY_PATH", "JUSPAY_JWE_PUBLIC_KEY")

	return &merchantConfigs
}

// readFile reads the contents of a file and returns it as a string
func readFile(path string) (string, error) {
	// Expand the path in case it contains environment variables
	expandedPath := os.ExpandEnv(path)

	// Resolve relative path to absolute
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return "", err
	}

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func PrintAllConfig() {
	allSettings := viper.AllSettings()

	// Get all environment variables
	envVars := os.Environ()

	// Print Viper settings
	logger.Info("Viper Settings:")
	logger.Info(strings.Repeat("=", 20))
	printMap(allSettings, 0)

	// Print Environment Variables
	logger.Info("\nEnvironment Variables:")
	logger.Info(strings.Repeat("=", 20))
	sort.Strings(envVars)
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			logger.Info("%s: %s\n", parts[0], parts[1])
		}
	}
}

func printMap(m map[string]interface{}, indent int) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	indentStr := strings.Repeat("  ", indent)

	for _, k := range keys {
		logger.Info("%s%s:", indentStr, k)
		v := m[k]
		switch v := v.(type) {
		case map[string]interface{}:
			printMap(v, indent+1)
		case []interface{}:
			printSlice(v, indent+1)
		default:
			logger.Info(" %v\n", v)
		}
	}
}

func printSlice(s []interface{}, indent int) {
	indentStr := strings.Repeat("  ", indent)
	for i, v := range s {
		logger.Info("%s%d:", indentStr, i)
		switch v := v.(type) {
		case map[string]interface{}:
			printMap(v, indent+1)
		case []interface{}:
			printSlice(v, indent+1)
		default:
			logger.Info(" %v\n", v)
		}
	}
}

func GetQueuePrefix() string {
	return viper.GetString("SQS_PREFIX")
}
