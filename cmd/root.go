package cmd

import (
	"log"
	"place-search/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "place-search",
	Short: "A high-performance geocoding search service",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", "config.yaml", "config file path")
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
}
