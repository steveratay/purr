package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"strings"
)

type UnwatchCmd struct {
	config *Config
}

func (h *UnwatchCmd) Help() string {
	return "`unwatch organisation/repo` - removes a github repo from the watch list"
}

// @todo check that purr have access / exists before removing repo
func (h *UnwatchCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "unwatch" {
		return true
	}
	return false
}

func (h *UnwatchCmd) GetResponse(msg *slack.MessageEvent) (string, error) {
	fields := strings.Fields(msg.Text)
	if len(fields) != 3 {
		return "", errors.New("usage: `unwatch organisation/repo`")
	}

	user, userFound := GetUser(msg.User)
	if !userFound || user.Profile.Email == "" {
		return "", fmt.Errorf("I can't let you do that %s, you are not in my list of approved users.", msg.User)
	}

	if h.config.AdminDomain != "" && !strings.HasSuffix(user.Profile.Email, h.config.AdminDomain) {
		return "", fmt.Errorf("I can't let you do that %s, you dont have an `%s` email address", user.Name, h.config.AdminDomain)
	}

	watchlist, err := h.config.WatchList(msg.Channel)
	if err != nil {
		return "", err
	}
	repoName := fields[2]
	newRepos := make([]string, 0)
	found := false
	for _, repo := range watchlist.Repositories {
		if repo == repoName {
			found = true
		} else {
			newRepos = append(newRepos, repo)
		}
	}
	watchlist.Repositories = newRepos
	if found {
		if err := h.config.SaveWatchList(watchlist); err != nil {
			logrus.Errorf("Unwatch: %s", err)
			return "", errors.New("I couldn't save that to by watch list :/")
		}
		return fmt.Sprintf("I've removed `%s` from the watchlist", repoName), nil
	}
	return fmt.Sprintf("I wasn't watching `%s` to begin with", repoName), nil
}
