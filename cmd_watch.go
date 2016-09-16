package main

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"strings"
)

type WatchCmd struct {
	config *Config
}

func (h *WatchCmd) Help() string {
	return "`watch organisation/repo` - add a github repo to the watch list"
}

func (h *WatchCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "watch" {
		return true
	}
	return false
}

// @todo check that purr have access / exists before adding repo
func (h *WatchCmd) GetResponse(msg *slack.MessageEvent) (string, error) {
	fields := strings.Fields(msg.Text)
	if len(fields) != 3 {
		return "", errors.New("usage: `watch organisation/repo`")
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
	found := false
	for _, repo := range watchlist.Repositories {
		if repo == repoName {
			found = true
		}
	}

	if !found {
		watchlist.Repositories = append(watchlist.Repositories, fields[2])

		if err := h.config.SaveWatchList(watchlist); err != nil {
			logrus.Errorf("Watch: %s", err)
			return "", errors.New("I couldn't save the change")
		}
		return fmt.Sprintf("I've added `%s` to the watchlist", repoName), nil
	}

	return fmt.Sprintf("I'm already watching `%s`", repoName), nil

}
