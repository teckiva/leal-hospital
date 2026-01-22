1. Centralized Configuration Management

  The file provides a single place to manage all application configuration including:
  - Database connections (MySQL database)

  Multi-Source Configuration Loading

  It supports loading configuration from multiple sources with priority

  Local Development:
  // Reads from .env files in the project
  viper.SetConfigName(".env")
  viper.SetConfigType("env")

  Environment Variables:
  viper.AutomaticEnv()  // Automatically maps env vars
4. Database Connection String Generation

  Automatically builds DSN (Data Source Name) strings:

  // MySQL Connection String
  config.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
      config.Database.Username,
      config.Database.Password,
      config.Database.Host,
      config.Database.Port,
      config.Database.Database,
  )