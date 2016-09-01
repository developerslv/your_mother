package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "your_mother",
	Short: "Your mother is a bot",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.your_mother.yaml)")

	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose")
	RootCmd.PersistentFlags().StringP("rpc_network", "", "tcp", "RPC network to use")
	RootCmd.PersistentFlags().StringP("rpc_address", "", "0.0.0.0:8080", "RPC address")
	RootCmd.PersistentFlags().StringP("sub_channel", "", "hn_top", "ably.io subscribe channel")
}

func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetEnvPrefix("YOUR_MOTHER")
	viper.SetConfigName(".your_mother") // name of config file (without extension)
	viper.AddConfigPath("$HOME")        // adding home directory as first search path
	viper.AutomaticEnv()                // read in environment variables that match

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
