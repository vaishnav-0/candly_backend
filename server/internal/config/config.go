package config

import (
	"encoding/json"
	"os"
)

var base_path string = "./"

type Config struct {
	Db struct {
		Host     string
		Username string
		Name     string
		Password string
	}
	Logging struct {
		// Enable console logging
		ConsoleLoggingEnabled bool

		// FileLoggingEnabled makes the framework log to a file
		// the fields below can be skipped if this value is false!
		FileLoggingEnabled bool
		// Directory to log to to when filelogging is enabled
		Directory string
		// Filename is the name of the logfile which will be placed inside the directory
		Filename string
		// MaxSize the max size in MB of the logfile before it's rolled
		MaxSize int
		// MaxBackups the max number of rolled files to keep
		MaxBackups int
		// MaxAge the max age in days to keep a logfile
		MaxAge int
	}
}

func GetConfig() Config {

	if path := os.Getenv("CANDLY_BASE"); path != "" {
		base_path = path
	}

	var config Config
	dat, err := os.ReadFile(base_path + "config/dev.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(dat, &config)

	return config
}
