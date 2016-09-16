package main

import (
	"github.com/nlopes/slack"
	"strings"
)

type HelpCmd struct{}

func (h *HelpCmd) Help() string {
	return "`help` - ask me for help"
}

func (h *HelpCmd) CanRespond(msg *slack.MessageEvent) bool {
	fields := strings.Fields(msg.Text)
	if len(fields) > 1 && fields[1] == "help" {
		return true
	}
	return false
}

func (h *HelpCmd) GetResponse(msg *slack.MessageEvent) (string, error) {
	var help string
	for _, cmd := range cmdList {
		help += "• " + cmd.Help() + "\n"
	}
	return help, nil
}
