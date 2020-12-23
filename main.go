package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kubqoa/tempo-cli/jira"
	"github.com/kubqoa/tempo-cli/tempo"
	"os"
	"regexp"
	"time"
)

var (
	log        = flag.String("log", "", "log a new workklog to Tempo")
	login      = flag.Bool("login", false, "preform login to Tempo")
	configFile = flag.String("config", func() string { dir, _ := os.UserConfigDir(); return dir + "/tempo-cli.toml" }(), "set the path to the config file")
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
	Credentials tempo.Credentials
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
		credentials, err := tempoAPI.Login()
		if err != nil {
			fmt.Println("there was an error logging in:", err)
			return 
		}

		conf.Credentials = credentials
		conf.writeConfig(*configFile)

		return
	}

	var c tempo.Credentials

	if conf.Credentials == c {
		fmt.Println("you need to login first")
		return
	}

	if conf.Credentials.ExpiresAt.Before(time.Now()) {
		fmt.Println("We need to refresh the Tempo access token. Hang on please.")
		credentials, err := tempoAPI.Refresh(conf.Credentials)
		if err != nil {
			fmt.Println("there was an error refreshing the access token", err)
			return 
		}
		conf.Credentials = credentials
		conf.writeConfig(*configFile)
	}

	fmt.Println(jiraAPI.GetCurrentUser())
}

func (c config) writeConfig(path string) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0660)
	if err != nil {
		fmt.Println("there was an error opening the configuration file for writing", err)
		return
	}
	err = toml.NewEncoder(file).Encode(c)
	if err != nil {
		fmt.Println("there was an error writing to the configuration file", err)
		return
	}
	return
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
