package cmd

import (
	"github.com/fatih/color"
	"github.com/kubqoa/tempo-cli/config"
	"github.com/kubqoa/tempo-cli/tempo"
)

func Login(tempoAPI tempo.TempoAPI, config config.Config) {
	credentials, err := tempoAPI.Login()
	if err != nil {
		color.Red("There was an error logging in: %v", err)
		return
	}

	config.Credentials = credentials
	if err := config.WriteConfig(); err != nil {
		color.Red("%v", err)
	}
}
