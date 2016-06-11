package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
	"github.com/spf13/cobra"
	"os"
)

var log = logrus.New()

func init() {
	sentryDSN := os.Getenv("SENTRY_DSN")

	if sentryDSN != "" {
		hook, err := logrus_sentry.NewSentryHook(sentryDSN, []logrus.Level{
			logrus.PanicLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		})

		if err == nil {
			log.Hooks.Add(hook)
		}
	}

	log.Level = logrus.DebugLevel
}

func main() {
	log.Info("Starting BOT")

	var nick, user, server, channel string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "your_mom",
		Short: "Your mother is a bot",
		Run: func(cmd *cobra.Command, args []string) {
			settings := IRCSettings{
				Nick:    nick,
				User:    user,
				Server:  server,
				Channel: channel,
			}

			c := NewIRCClient(settings, NewWeather(), NewHackersNews())

			c.Start()
		},
	}

	cmd.Flags().StringVarP(&nick, "nick", "n", "Your_mother_BOT", "Bot name")
	cmd.Flags().StringVarP(&user, "user", "u", "Your_mother_BOT", "Bot user name")
	cmd.Flags().StringVarP(&server, "server", "s", "irc.freenode.net:6667", "Server to which connect to")
	cmd.Flags().StringVarP(&channel, "channel", "c", "#developerslv", "Channel to which connect to")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Log irc output")

	cmd.Execute()
}
