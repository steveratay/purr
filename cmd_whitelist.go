package main

import (
	"errors"
	"fmt"
	"github.com/nlopes/slack"
	"strings"
)

type WhiteListCmd struct {
	config *Config
}

func (h *WhiteListCmd) Help() string {
	help := "`whitelist add/remove <github_username>` - add github user to the whitelist"
	return help
}

func (h *WhiteListCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "whitelist" {
		return true
	}
	return false
}

// @todo check that purr have access / exists before adding / removing user
func (h *WhiteListCmd) GetResponse(msg *slack.MessageEvent) (string, error) {

	fields := strings.Fields(msg.Text)
	if len(fields) != 4 {
		return "", errors.New("usage: whitelist add/remove")
	}

	user, userFound := GetUser(msg.User)
	if !userFound || user.Profile.Email == "" {
		return "", fmt.Errorf("I can't let you do that %s, you are not in my list of approved users.", msg.User)
	}
	if h.config.AdminDomain != "" && !strings.HasSuffix(user.Profile.Email, h.config.AdminDomain) {
		return "", fmt.Errorf("I can't let you do that %s, you dont have an `%s` email address", user.Name, h.config.AdminDomain)
	}

	watchList, err := h.config.WatchList(msg.Channel)
	if err != nil {
		return "", err
	}

	if fields[2] == "add" {
		watchList.UserWhiteList = append(watchList.UserWhiteList, fields[3])
		err := h.config.SaveWatchList(watchList)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("I've added `%s` to the whitelist", fields[3]), nil
	} else if fields[2] == "remove" {
		newList := make([]string, 0)
		for _, user := range watchList.UserWhiteList {
			if user != fields[3] {
				newList = append(newList, user)
			}
		}
		watchList.UserWhiteList = newList
		err := h.config.SaveWatchList(watchList)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("I've removed `%s` from the whitelist", fields[3]), nil
	}

	return "", errors.New("usage: whitelist add/remove <username>")

}
