package cmd

import (
	"log"
	"os"

	"github.com/0xAFz/ff/internal/udp"
	"github.com/spf13/cobra"
)

var (
	addr string
	dest string
)

var rootCmd = &cobra.Command{
	Use:   "ff",
	Short: "Fast Packet Forwarder",
	Run: func(_ *cobra.Command, _ []string) {
		if err := udp.ListenForwarder(addr, dest); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&addr, "addr", "", "Listen Address (127.0.0.1:8080)")
	rootCmd.Flags().StringVar(&dest, "dest", "", "Destination Address (127.0.0.1:8081)")

	rootCmd.MarkFlagRequired("addr")
	rootCmd.MarkFlagRequired("dest")
}
