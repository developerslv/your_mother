package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dainis/your_mother/bot"
	"github.com/spf13/cobra"
)

var ircCmd = &cobra.Command{
	Use:   "irc",
	Short: "Your moms rpc server",
	Run: func(cmd *cobra.Command, args []string) {
		settings := &bot.IRCSettings{}
		settings.Nick, _ = cmd.Flags().GetString("nick")
		settings.User, _ = cmd.Flags().GetString("user")
		settings.Channel, _ = cmd.Flags().GetString("channel")
		settings.Server, _ = cmd.Flags().GetString("irc_server")
		settings.RPCNetwork, _ = cmd.Flags().GetString("rpc_network")
		settings.RPCAddress, _ = cmd.Flags().GetString("rpc_address")
		settings.Debug, _ = cmd.Flags().GetBool("verbose")

		irc := bot.NewIRCClient(settings)

		log.Debug("Will start IRC loop")

		err := irc.Start()

		if err != nil {
			log.WithError(err).Panic("Failed to start irc client")
		}
	},
}

func init() {
	RootCmd.AddCommand(ircCmd)

	ircCmd.Flags().StringP("nick", "n", "Your_mom_BOT", "Nick to use")
	ircCmd.Flags().StringP("user", "u", "Your_mom_BOT", "User name to use")
	ircCmd.Flags().StringP("channel", "c", "#developerslv", "Channel to join")
	ircCmd.Flags().StringP("irc_server", "s", "irc.freenode.net:6667", "Server to connect to")
}
