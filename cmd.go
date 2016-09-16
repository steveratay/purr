package main

import (
	"github.com/nlopes/slack"
)

type Command interface {
	// will return help and usage information when users ask purr about help
	Help() string
	CanRespond(msg *slack.MessageEvent) bool
	GetResponse(msg *slack.MessageEvent) (string, error)
}
