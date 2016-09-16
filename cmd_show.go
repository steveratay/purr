package main

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"strings"
)

type ShowCmd struct {
	config *Config
}

func (h *ShowCmd) Help() string {
	return "`show` - show all open pull requests"
}

func (h *ShowCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "show" {
		return true
	}
	return false
}

func (h *ShowCmd) GetResponse(msg *slack.MessageEvent) (string, error) {

	fields := strings.Fields(msg.Text)
	if len(fields) != 2 {
		return "", errors.New("usage: show")
	}
	watchlist, err := h.config.WatchList(msg.Channel)
	if err != nil {
		return "", err
	}

	// these function will return channels that will emit a list of pull requests
	// on channels and close the channel when they are done
	gitHubPRs := trawlGitHub(h.config, watchlist)
	//gitLabPRs := trawlGitLab(conf)

	// Merge the in channels into of channel and close it when the inputs are done
	//prs := merge(gitHubPRs, gitLabPRs)
	prs := merge(gitHubPRs)

	// filter out pull requests that we don't want to send
	filteredPRs := filter(watchlist, prs)

	// format takes a channel of pull requests and returns a message that groups
	// pull request into repos and formats them into a slack friendly format
	message := format(filteredPRs)

	// Output what slack will send if we are in debug mode
	if debug {
		logrus.Debugf("Final message:\n%s", message)
	}

	return message.String(), nil
}
