package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "lightshow",
		Short: "Lightshow app allows syncing Spotify songs and Hue lights through the use of custom \"maps\".",
	}
)

func init() {
	viper.SetEnvPrefix("lightshow")
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringP("bridge-host", "b", "", "Bridge Host")
	rootCmd.PersistentFlags().IntP("bridge-port", "p", 2100, "Bridge Port (for UDP DTLS comms)")
	rootCmd.PersistentFlags().StringP("bridge-user", "u", "", "Bridge Username (pre-registered)")
	rootCmd.PersistentFlags().StringP("bridge-key", "k", "", "Bridge Client Key (pre-registered)")
	rootCmd.PersistentFlags().IntP("entertainment-group", "g", 0, "Entertainment Group on Bridge to use (pre-created)")

	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("bridge-host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("bridge-port"))
	viper.BindPFlag("user", rootCmd.PersistentFlags().Lookup("bridge-user"))
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("bridge-key"))
	viper.BindPFlag("group", rootCmd.PersistentFlags().Lookup("entertainment-group"))
}

func main() {
	fmt.Fprintln(os.Stderr, rootCmd.Execute())
}
