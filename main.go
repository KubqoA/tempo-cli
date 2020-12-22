package main

import (
	"errors"
	"flag"
	"fmt"
	"regexp"
	"github.com/BurntSushi/toml"
	"github.com/kubqoa/tempo-cli/jira"
	"github.com/kubqoa/tempo-cli/tempo"
)

var (
	login      = flag.Bool("login", false, "preform login to tempo")
	configFile = flag.String("config", "~/.tempo-cli.toml", "set the path to the config file")
)

type config struct {
	JiraUrl string
	Tempo   struct {
		ClientId     string
		ClientSecret string
	}
	Jira struct {
		Email    string
		ApiToken string
	}
}

func main() {
	var conf config
	flag.Parse()

	if _, err := toml.DecodeFile(*configFile, &conf); err != nil {
		fmt.Println("error parsing configuration file:", err)
		return
	}

	if err := conf.validateConfig(); err != nil {
		fmt.Println("invalid configuration file:", err)
		return
	}

	jiraAPI := jira.JiraAPI{
		JiraUrl:  conf.JiraUrl,
		Email:    conf.Jira.Email,
		ApiToken: conf.Jira.ApiToken,
	}

	tempoAPI := tempo.TempoAPI{
		ClientId:     conf.Tempo.ClientId,
		ClientSecret: conf.Tempo.ClientSecret,
		JiraUrl:      conf.JiraUrl,
	}

	if *login {
		code, err := tempoAPI.Login()
		fmt.Printf("got access token: %v with error %v\n", code.AccessToken, err)
	} else {
		fmt.Println(jiraAPI.GetCurrentUser())
	}
}

func (c config) validateConfig() error {
	if !regexp.MustCompile(`http[s]?://.*`).MatchString(c.JiraUrl) {
		return errors.New("jiraUrl must be in format http(s)://url-of-jira")
	}
	if c.Tempo.ClientId == "" {
		return errors.New("tempo.clientId must not be empty")
	}
	if c.Tempo.ClientSecret == "" {
		return errors.New("tempo.clientSecret must not be empty")
	}
	if !regexp.MustCompile(`.*@.*`).MatchString(c.Jira.Email) {
		return errors.New("jira.email is not a valid email")
	}
	if c.Jira.ApiToken == "" {
		return errors.New("jira.apiToken must not be empty")
	}
	return nil
}