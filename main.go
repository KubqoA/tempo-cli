package main

import (
	"flag"
	"github.com/fatih/color"
	"github.com/kubqoa/tempo-cli/cmd"
	"github.com/kubqoa/tempo-cli/config"
	"github.com/kubqoa/tempo-cli/jira"
	"github.com/kubqoa/tempo-cli/tempo"
	"os"
)

var (
	log        = flag.String("log", "", "log a new workklog to Tempo")
	login      = flag.Bool("login", false, "preform login to Tempo")
	configFile = flag.String("config", func() string { dir, _ := os.UserConfigDir(); return dir + "/tempo-cli.toml" }(), "set the path to the config file")
)

func main() {
	flag.Parse()

	config, err := config.MakeConfig(*configFile)
	if err != nil {
		color.Red("Error parsing configuration file: %v", err)
		return
	}

	jiraAPI := jira.JiraAPI{
		JiraUrl:  config.JiraUrl,
		Email:    config.Jira.Email,
		ApiToken: config.Jira.ApiToken,
	}

	tempoAPI := tempo.TempoAPI{
		ClientId:     config.Tempo.ClientId,
		ClientSecret: config.Tempo.ClientSecret,
		JiraUrl:      config.JiraUrl,
	}

	if *login {
		cmd.Login(tempoAPI, config)
		return
	}

	if err := config.ValidateCredentials(); err != nil {
		color.Red("%v", err)
	}

	cmd.Renew(tempoAPI, config)

	user, _ := jiraAPI.GetCurrentUser()

	color.White("%v", user)
}
