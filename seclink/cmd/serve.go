/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"seclink/api"
	"seclink/db"
	"seclink/log"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts a seclink API server",
	Long:  `Uses the config file seclink.yaml for settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		Serve()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func Serve() error {
	l := log.Get()
	db := db.NewSeclinkDb()
	err := db.Start(false, false)
	if err != nil {
		l.Error().Err(err).Msg("An error occurred opening the database")
		return err
	}
	api := api.NewSeclinkApi(db)
	err = api.Start()
	if err != nil {
		l.Error().Err(err).Msg("An error occurred starting the API")
		return err
	}
	return nil

}
