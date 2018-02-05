package config

import "os"

// Get returns a configuration value given a particular key
func Get(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	defValue, ok := defaults[key]
	if !ok {
		return ""
	}

	return defValue
}