package config

import (
	"capsulecd/pkg/utils"
	"encoding/base64"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

//Viper uses the following precedence order. Each item takes precedence over the item below it:
// explicit call to Set
// flag
// env
// config
// key/value store
// default
var v = viper.New()
var initialized = false

func Init() {
	v = viper.New()
	//set defaults
	v.SetDefault("package_type", "default")
	v.SetDefault("scm", "default")
	v.SetDefault("runner", "default")
	v.SetDefault("engine_version_bump_type", "patch")
	v.SetDefault("chef_supermarket_type", "Other")

	//set the default system config file search path.
	//if you want to load a non-standard location system config file (~/capsule.yml), use ReadConfig
	//if you want to load a repo specific config file, use ReadConfig
	v.SetConfigType("yaml")
	v.SetConfigName("capsule")
	v.AddConfigPath("$HOME/")

	//configure env variable parsing.
	v.SetEnvPrefix("CAPSULE")
	v.AutomaticEnv()

	//CLI options will be added via the `Set()` function

	initialized = true
}

func IsInitialized() bool {
	return initialized
}

func ReadConfig(configFilePath string) {

	if !utils.FileExists(configFilePath) {
		log.Print("The configuration file could not be found. Skipping")
		return
	}

	log.Printf("Loading configuration file: %s", configFilePath)

	config_data, err := os.Open(configFilePath)
	if err != nil {
		log.Printf("Error reading configuration file: %s", err)
		return
	}
	v.MergeConfig(config_data)
}

func Set(key string, value interface{}) {
	v.Set(key, value)
}

func AllSettings() map[string]interface{} {
	return v.AllSettings()
}

func IsSet(key string) bool {
	return v.IsSet(key)
}

func Get(key string) interface{} {
	return v.Get(key)
}

func GetBool(key string) bool {
	return v.GetBool(key)
}

func GetDuration(key string) time.Duration {
	return v.GetDuration(key)
}

func GetFloat64(key string) float64 {
	return v.GetFloat64(key)
}

func GetInt(key string) int {
	return v.GetInt(key)
}

func GetString(key string) string {
	return v.GetString(key)
}

func GetStringMap(key string) map[string]interface{} {
	return v.GetStringMap(key)
}

func GetStringMapString(key string) map[string]string {
	return v.GetStringMapString(key)
}

func GetStringMapStringSlice(key string) map[string][]string {
	return v.GetStringMapStringSlice(key)
}

func GetStringSlice(key string) []string {
	return v.GetStringSlice(key)
}

func GetTime(key string) time.Time {
	return v.GetTime(key)
}

func GetBase64Decoded(key string) string {
	if len(v.GetString(key)) > 0 {
		key, err := base64.StdEncoding.DecodeString(v.GetString(key))
		if err != nil {
			log.Print("Could not decode chef_supermarket_key")
		}
		return string(key)
	}
	return ""
}
