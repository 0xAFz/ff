package cmd

import (
	"log"
	"os"
	"runtime"

	"github.com/0xAFz/ff/internal/udp"
	"github.com/spf13/cobra"
)

var (
	laddr string
	raddr string
)

var rootCmd = &cobra.Command{
	Use:   "ff",
	Short: "Fast Packet Forwarder",
	Run: func(_ *cobra.Command, _ []string) {
		runtime.GOMAXPROCS(2)

		forwarder := udp.NewForwarder()

		if err := forwarder.Init(laddr, raddr); err != nil {
			log.Printf("failed to resolve addresses: %v", err)
		}

		if err := forwarder.Listen(); err != nil {
			log.Printf("error when listening: %v", err)
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
	rootCmd.Flags().StringVar(&laddr, "laddr", "", "Listen Address (127.0.0.1:8080)")
	rootCmd.Flags().StringVar(&raddr, "raddr", "", "Destination Address (127.0.0.1:8081)")

	rootCmd.MarkFlagRequired("laddr")
	rootCmd.MarkFlagRequired("raddr")
}
