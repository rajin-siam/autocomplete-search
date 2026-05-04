package config

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Elasticsearch ESConfig     `yaml:"Elasticsearch"`
	Server        ServerConfig `yaml:"server"`
	Search        SearchConfig `yaml:"search"`
}

type ESConfig struct {
	Host  string `yaml:"host"`
	Index string `yaml:"index"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type SearchConfig struct {
	MinChars   int    `yaml:"min_chars"`
	Fuzziness  string `yaml:"fuzziness"`
	MaxResults int    `yaml:"max_results"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	expanded := os.ExpandEnv(string(data))

	cfg := &Config{}
	if err := yaml.Unmarshal([]byte(expanded), cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
