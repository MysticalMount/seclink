/*
Copyright © 2024 Jay <jayaya369@proton.me>
*/
package cmd

import (
	"os"
	"seclink/log"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	cliConfig SCliConfig
	l         zerolog.Logger
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
	cobra.OnInitialize(initLog, initConfig, printConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is seclink.yaml in launch directory)")
	rootCmd.PersistentFlags().IntVarP(&cliConfig.LogLevel, "vervose", "v", 1, "sets the log level, -1 for trace, 0 for debug, 1 for info")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initLog creates the logger, the verbosity is set on the command line via global flag -v
func initLog() {
	log.InitLog(cliConfig.LogLevel)
	l = log.Get()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

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

}

// printConfig prints the config to the output
func printConfig() {
	l.Info().
		Int("LogLevel", cliConfig.LogLevel).
		Int("Port", viper.GetInt("server.port")).
		Int("AdminPort", viper.GetInt("server.adminport")).
		Str("DataPath", viper.GetString("server.datapath")).
		Str("ExternalURL", viper.GetString("server.datapath")).
		Str("DefaultTTL", viper.GetDuration("links.defaultttl").String()).
		Msg("Printing configuration")
}
