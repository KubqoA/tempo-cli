package config

import (
	"errors"
	"github.com/BurntSushi/toml"
	"github.com/kubqoa/tempo-cli/tempo"
	"os"
	"regexp"
)

type Config struct {
	path    string
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

func MakeConfig(path string) (config Config, err error) {
	if _, err = toml.DecodeFile(path, &config); err != nil {
		return
	}
	if err = config.validateConfig(); err != nil {
		return
	}
	config.path = path
	return
}

func (c Config) WriteConfig() error {
	file, err := os.OpenFile(c.path, os.O_WRONLY, 0660)
	if err != nil {
		return errors.New("There was an error opening the congfiguration file for writing: " + err.Error())
	}
	err = toml.NewEncoder(file).Encode(c)
	if err != nil {
		return errors.New("There was an error writing to the configuration file: " + err.Error())
	}
	return nil
}

func (c Config) ValidateCredentials() error {
	var credentials tempo.Credentials

	if c.Credentials == credentials {
		return errors.New("You are not logged in. Log in first by running the program with -login")
	}

	return nil
}

func (c Config) validateConfig() error {
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
