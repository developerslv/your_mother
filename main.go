package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var nick, user, server, channel string

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

	cmd.Execute()
}
