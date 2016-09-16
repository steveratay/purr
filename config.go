package main

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
)

type WatchList struct {
	// slack channel id
	Channel       string   `json:"channel"`
	Repositories  []string `json:"repositories"`
	UserWhiteList []string `json:"user_whitelist"`
}

// Config contains the settings from the user
type Config struct {
	filePath    string
	dataDir     string
	GitHubToken string `json:"github_token"`
	GitLabToken string `json:"gitlab_token"`
	GitlabURL   string `json:"gitlab_url"`
	SlackToken  string `json:"slack_token"`
	AdminDomain string `json:"admin_domain"`
}

func (c *Config) WatchList(channel string) (*WatchList, error) {
	w := &WatchList{
		Channel: channel,
	}
	file, err := ioutil.ReadFile(c.dataDir + "/" + channel)
	if os.IsNotExist(err) {
		return w, nil
	} else if err != nil {
		return nil, fmt.Errorf("Error during watchlist read: %s", err)
	}
	if err := json.Unmarshal(file, &w); err != nil {
		return nil, fmt.Errorf("Error during watchlist unmarshalling: %s", err)
	}
	return w, nil
}

func (c *Config) SaveWatchList(w *WatchList) error {
	content, err := json.MarshalIndent(w, "", "\t")
	if err != nil {
		logrus.Errorf("Error during watchlist save: %s", err)
		return err
	}
	return ioutil.WriteFile(c.dataDir+"/"+w.Channel, content, 0600)
}

func newConfig(filePath string) (*Config, error) {

	c := &Config{}

	if filePath != "" {
		file, err := ioutil.ReadFile(filePath)
		if err != nil {
			return c, fmt.Errorf("Error during config read: %s", err)
		}
		if err := json.Unmarshal(file, &c); err != nil {
			return c, fmt.Errorf("Error during config read: %s", err)
		}
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		c.GitHubToken = os.Getenv("GITHUB_TOKEN")
	}
	if os.Getenv("GITLAB_TOKEN") != "" {
		c.GitLabToken = os.Getenv("GITLAB_TOKEN")
	}
	if os.Getenv("GITLAB_URL") != "" {
		c.GitlabURL = os.Getenv("GITLAB_URL")
	}
	if os.Getenv("SLACK_TOKEN") != "" {
		c.SlackToken = os.Getenv("SLACK_TOKEN")
	}

	c.filePath = filePath
	c.dataDir = "/tmp/purr"

	if c.dataDir != "" {
		stat, err := os.Stat(c.dataDir)
		if os.IsNotExist(err) {
			logrus.Warnf("Creating data directory '%s'", c.dataDir)
			err := os.Mkdir(c.dataDir, 0770)
			return c, err
		} else if err != nil {
			return c, err
		}
		if !stat.IsDir() {
			return c, fmt.Errorf("data directory '%s' is not a directory", c.dataDir)
		}
	}
	return c, nil
}

func (c *Config) validate() []error {
	var errors []error
	if c.GitHubToken == "" {
		errors = append(errors, fmt.Errorf("GitHub token cannot be empty"))
	}
	if c.SlackToken == "" {
		errors = append(errors, fmt.Errorf("Slack token cannot be empty"))
	}
	return errors
}

func configHelp() {
	fmt.Fprintln(os.Stderr, "\npurr requrires configuration to be either in a config file or set the ENV")

	fmt.Fprintln(os.Stderr, "\nThe configuration file (--config) looks like this:")

	exampleConfig := &Config{
		GitHubToken: "secret_token",
		GitLabToken: "secret_token",
		GitlabURL:   "https://www.example.com",
		SlackToken:  "secret_token",
	}
	b, err := json.MarshalIndent(exampleConfig, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
	}
	fmt.Fprintf(os.Stderr, "\n%s\n\n", b)

	fmt.Fprint(os.Stderr, "The above configuration can be overridden with ENV variables:\n\n")
	fmt.Fprintln(os.Stderr, " * GITHUB_TOKEN")
	fmt.Fprintln(os.Stderr, " * GITLAB_TOKEN")
	fmt.Fprintln(os.Stderr, " * GITLAB_URL")
	fmt.Fprintln(os.Stderr, " * SLACK_TOKEN")
	fmt.Fprintln(os.Stderr, " * SLACK_CHANNEL")
}
