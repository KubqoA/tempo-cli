package cmd

import (
	"github.com/fatih/color"
	"github.com/kubqoa/tempo-cli/config"
	"github.com/kubqoa/tempo-cli/tempo"
	"time"
)

func Renew(tempoAPI tempo.TempoAPI, config config.Config) {
	if config.Credentials.ExpiresAt.Before(time.Now()) {
		color.Yellow("We need to refresh the Tempo access token. Hang on please.")
		credentials, err := tempoAPI.Refresh(config.Credentials)
		if err != nil {
			color.Red("There was an error refreshing the access token %v", err)
			return
		}
		config.Credentials = credentials
		if err := config.WriteConfig(); err != nil {
			color.Red("%v", err)
			return
		}
		color.Green("Successfully renewed")
	}
}
