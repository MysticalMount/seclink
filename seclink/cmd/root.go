/*
Copyright © 2024 Jay <jayaya369@proton.me>
*/
package cmd

import (
	"os"
	"seclink/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile  string
	filePath *string
	ttl      *int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "seclink",
	Short: "Generates secure time based links for a level of secure sharing of individual files over public facing http",
	Long: `Generates secure time based links for a level of secure sharing of
	individual files over public facing http`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, printConfig, setLogLevel)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is seclink.yaml in launch directory)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	l := log.Get()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in app directory
		viper.AddConfigPath("/seclink")
		viper.SetConfigType("yaml")
		viper.SetConfigName("seclink")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		l.Info().Str("ConfigFile", viper.ConfigFileUsed()).Msg("Config file used")
	}

	// Set defaults to specified
	if *ttl == 0 {
		ttlVal := viper.GetInt("links.defaultttl")
		ttl = &ttlVal
	}
}

// printConfig prints the config to the output
func printConfig() {
	l := log.Get()

	l.Info().
		Int("LogLevel", viper.GetInt("server.loglevel")).
		Int("Port", viper.GetInt("server.port")).
		Str("DataPath", viper.GetString("server.datapath")).
		Int("DefaultTTL", viper.GetInt("links.defaultttl")).
		Msg("Printing configuration")

	log.SetLevel(viper.GetInt("server.loglevel"))
}

// setLogLevel sets the loglevel, this has to be done after the config has loaded
func setLogLevel() {
	log.SetLevel(viper.GetInt("server.loglevel"))
}
