package config

import "os"

type Config struct {
	SQLitePath string
}

func Load() Config {
	return Config{
		SQLitePath: os.Getenv("SQLITE_PATH"),
	}
}
