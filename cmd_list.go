package main

import (
	"errors"
	"fmt"
	"github.com/nlopes/slack"
	"strings"
)

type ListCmd struct {
	config *Config
}

func (h *ListCmd) Help() string {
	return "`list` - list all watched repositories"
}

func (h *ListCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "list" {
		return true
	}
	return false
}

func (h *ListCmd) GetResponse(msg *slack.MessageEvent) (string, error) {
	fields := strings.Fields(msg.Text)
	if len(fields) != 2 {
		return "", errors.New("usage: list")
	}
	watchlist, err := h.config.WatchList(msg.Channel)
	if err != nil {
		return "", err
	}

	found := len(watchlist.Repositories) > 0
	if !found {
		return "I'm currently not watching any repositories\n", nil
	}

	response := ""
	for _, repoName := range watchlist.Repositories {
		response += fmt.Sprintf("• %s\n", repoName)
	}
	return "I'm currently watching:\n" + response, nil
}
