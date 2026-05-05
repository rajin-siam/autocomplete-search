package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Elasticsearch ESConfig
	Server        ServerConfig
	Search        SearchConfig
	Database      DatabaseConfig
}

type ESConfig struct {
	Host  string
	Index string
}

type ServerConfig struct {
	Port int
}

type SearchConfig struct {
	MinChars   int `mapstructure:"min_chars"`
	Fuzziness  string
	MaxResults int `mapstructure:"max_results"`
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func Load() (*Config, error) {
	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()

	viper.BindEnv("elasticsearch.host", "ES_HOST")
	viper.BindEnv("elasticsearch.index", "ES_INDEX")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("search.min_chars", "SEARCH_MIN_CHARS")
	viper.BindEnv("search.fuzziness", "SEARCH_FUZZINESS")
	viper.BindEnv("search.max_results", "SEARCH_MAX_RESULTS")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
