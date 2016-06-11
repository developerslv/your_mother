package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dainis/your_mother/bot"
	"github.com/dainis/your_mother/bot/suppliers"
	"github.com/spf13/cobra"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

var rpcCmd = &cobra.Command{
	Use:   "rpc",
	Short: "Your mother is a RPC",
	Run: func(cmd *cobra.Command, args []string) {
		news := suppliers.NewHackersNews()
		news.BackgroundNewLoop()

		rpcSrv := bot.NewRPCServer(suppliers.NewWeather(), news)
		err := rpc.Register(rpcSrv)

		if err != nil {
			log.WithError(err).Panic("Failed to register rpc server")
		}

		rpc.HandleHTTP()

		t, addr := cmd.Flag("rpc_network").Value.String(), cmd.Flag("rpc_address").Value.String()

		if t == "unix" {
			if err := os.Remove(addr); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"path": addr,
				}).Debug("Failed to remove unix socket")
			}
		}

		listener, err := net.Listen(t, addr)

		if err != nil {
			log.WithError(err).Panicf("Failed to start listening %s %s", t, addr)
		}

		log.Debugf("Listening to %s %s", t, addr)

		http.Serve(listener, nil)
	},
}

func init() {
	RootCmd.AddCommand(rpcCmd)
}
