package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strings"
)

var slackUsers map[string]slack.User

const (
	// BANNER is what is printed for help/info output
	BANNER = "purr - %s\n"
	// VERSION is the binary version.
	VERSION = "v0.5.0-alpha"
)

var (
	configFile string
	debug      bool
	cmdList    []Command
)

func init() {
	flag.StringVar(&configFile, "config", "", "Read config from FILE")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		flag.PrintDefaults()
		configHelp()
	}
	flag.Parse()

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	slackUsers = make(map[string]slack.User)
}

func main() {
	conf, err := newConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	if errors := conf.validate(); len(errors) > 0 {
		buf := &bytes.Buffer{}
		for i := range errors {
			fmt.Fprintln(buf, errors[i].Error())
		}
		usageAndExit(buf.String(), 1)
	}
	cmdList = make([]Command, 0)
	cmdList = append(cmdList, &HelpCmd{})
	cmdList = append(cmdList, &ShowCmd{config: conf})
	cmdList = append(cmdList, &WatchCmd{config: conf})
	cmdList = append(cmdList, &UnwatchCmd{config: conf})
	cmdList = append(cmdList, &ListCmd{config: conf})
	cmdList = append(cmdList, &WhiteListCmd{config: conf})
	connectToSlack(conf)
}

func AddUser(user slack.User) {
	slackUsers[user.ID] = user
	if user.Profile.Email != "" {
		logrus.Debugf("added user %s %s", user.ID, user.Profile.Email)
	} else {
		logrus.Debugf("added user %s - %s", user.ID, user.Name)
	}
}

func GetUser(id string) (slack.User, bool) {
	user, found := slackUsers[id]
	return user, found
}

func connectToSlack(conf *Config) {
	api := slack.New(conf.SlackToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(false)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	var purr *slack.UserDetails

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
			// Ignore hello

			case *slack.ConnectedEvent:
				purr = ev.Info.User

				for _, user := range ev.Info.Users {
					AddUser(user)
				}
				logrus.Debugf("Added %d users", len(slackUsers))

				logrus.Printf("Connected to slack as %s", ev.Info.User.Name)
				logrus.Debugf("Connection counter: %d", ev.ConnectionCount)

			case *slack.UserChangeEvent:
				logrus.Debugf("UserChangeEvent %+v", ev)

			case *slack.PresenceChangeEvent:
				logrus.Debugf("PresenceChangeEvent %+v", ev)

			case *slack.MessageEvent:
				if !strings.HasPrefix(ev.Text, "<@"+purr.ID+">") {
					break
				}
				go handleMessageEvent(ev, rtm)

			case *slack.LatencyReport:
				logrus.Debugf("Current latency: %s", ev.Value)

			case *slack.RTMError:
				logrus.Errorf("Error: %s", ev.Error())

			case *slack.InvalidAuthEvent:
				logrus.Errorf("Invalid credentials")
				break Loop

			default:
				// Ignore other events..
				// fmt.Printf("Unexpected: %v\n", msg.Data)
			}
		}
	}
}

func handleMessageEvent(ev *slack.MessageEvent, rtm *slack.RTM) {

	var commandFound bool
	var response string
	var err error
	for _, cmd := range cmdList {
		if !cmd.CanRespond(ev) {
			continue
		}
		commandFound = true
		logrus.Debugf("%s will respond to message '%s' from %s in channel %s", reflect.TypeOf(cmd), ev.Text, ev.User, ev.Channel)
		response, err = cmd.GetResponse(ev)
		if err == nil {
			logrus.Debug(response)
		} else {
			logrus.Errorf("Error during %s.GetResponse: %s", reflect.TypeOf(cmd), err)
			response = err.Error()
		}
		break
	}

	params := slack.PostMessageParameters{
		AsUser:    false,
		Username:  "purr",
		IconEmoji: ":purr:",
	}

	if response != "" {
		const maxLines = 30
		lines := strings.Split(response, "\n")
		lineBuffer := make([]string, maxLines)
		for i := range lines {
			lineBuffer = append(lineBuffer, lines[i])
			if len(lineBuffer) == cap(lineBuffer) || i+1 == len(lines) {
				msg := strings.Join(lineBuffer, "\n")
				if msg != "" {
					_, _, err := rtm.PostMessage(ev.Channel, msg, params)
					if err != nil {
						logrus.Errorf("Slack: %s", err)
					}
				}
				lineBuffer = make([]string, cap(lineBuffer))
			}
		}
	}

	if !commandFound {
		what := []string{"eh?", "um?", "¿que?", "meow?"}

		rtm.SendMessage(rtm.NewOutgoingMessage(what[rand.Intn(len(what))], ev.Channel))
	}
}

// postToSlack will post the message to Slack. It will divide the message into smaller message if
// it's more than 30 lines long due to a max message size limitation enforced by the Slack API
//func postToSlack(conf *Config, message fmt.Stringer) {
//
//	const maxLines = 30
//
//	if message.String() == "" {
//		return
//	}
//
//	client := slack.New(conf.SlackToken)
//	opt := &slack.ChatPostMessageOpt{
//		AsUser:    false,
//		Username:  "purr",
//		IconEmoji: ":purr:",
//	}
//
//	// Don't send to large messages, send a new message per 40 new lines
//}

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
