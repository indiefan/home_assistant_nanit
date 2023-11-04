package utils

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// TODO: We should probably use some library if there is need for additional functionality

// EnvVarStr - retrieves value of string environment variable, while applying default
func EnvVarStr(varName string, defaultValue string) string {
	value := os.Getenv(varName)

	if value == "" {
		return defaultValue
	}

	return value
}

// EnvVarReqStr - retrieves value of string environment variable, fails if it is not present or empty
func EnvVarReqStr(varName string) string {
	value := EnvVarStr(varName, "")

	if value == "" {
		log.Fatal().Msgf("Missing environment variable %v", varName)
	}

	return value
}

// EnvVarBool - retrieves value of boolean environment variable, fails if variable contains non-boolean value
func EnvVarBool(varName string, defaultValue bool) bool {
	value := EnvVarStr(varName, "")
	if value == "true" {
		return true
	} else if value == "false" {
		return false
	} else if value == "" {
		return defaultValue
	}

	log.Fatal().Msgf("Unexpected value for boolean environment variable %v (allowed values true, false)", varName)
	return false
}

// EnvVarSeconds - retrieves value of environment variable reperesenting duration in seconds, fails if variable non-parseable values
func EnvVarSeconds(varName string, defaultValue time.Duration) time.Duration {
	valueStr, found := os.LookupEnv(varName)

	if !found {
		return defaultValue
	}

	valueInt, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Fatal().Msgf("Unexpected value %v for environment variable %v", valueStr, varName)
	}

	value := time.Duration(valueInt) * time.Second

	return value
}

// LoadDotEnvFile - Loads environment variables from .env file in the current working directory (if found)
func LoadDotEnvFile() {
	absFilepath, filePathErr := filepath.Abs(".env")
	if filePathErr != nil {
		log.Fatal().Str("path", absFilepath).Err(filePathErr).Msg("Unable to retrieve absolute file path")
	}

	// loads values from .env into the system
	if err := godotenv.Load(absFilepath); err != nil {
		log.Info().Str("path", absFilepath).Msg("No .env file found. Using only environment variables")
	} else {
		log.Info().Str("path", absFilepath).Msg("Additional environment variables loaded from .env file")
	}
}
