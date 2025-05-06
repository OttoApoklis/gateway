package main

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DB       int    `yaml:"db"`
		Password string `yaml:"password"`
	} `yaml:"redis"`

	RateLimiter struct {
		WindowSize  int    `yaml:"window_size"`
		MaxRequests int    `yaml:"max_requests"`
		KeyPrefix   string `yaml:"key_prefix"`
	} `yaml:"rate_limiter"`

	Server struct {
		Port int `yaml:"port"`
	}
}

func LoadConfig(path string) *Config {
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return &cfg
}
