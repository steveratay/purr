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
	help := "`whitelist add <github_username>` - add github user to the whitelist\n"
	help += "`whitelist remove <github_username>` - remove github user to the whitelist"
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
	if len(fields) != 2 {
		return "", errors.New("usage: show")
	}
	_, err := h.config.WatchList(msg.Channel)
	if err != nil {
		return "", err
	}

	user, userFound := GetUser(msg.User)
	if !userFound || user.Profile.Email == "" {
		return "", fmt.Errorf("I can't let you do that %s, you are not in my list of approved users.", msg.User)
	}

	if h.config.AdminDomain != "" && !strings.HasSuffix(user.Profile.Email, h.config.AdminDomain) {
		return "", fmt.Errorf("I can't let you do that %s, you dont have an `%s` email address", user.Name, h.config.AdminDomain)
	}

	return "not implemented yet", nil
}
