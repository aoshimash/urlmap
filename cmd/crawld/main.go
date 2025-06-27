package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "crawld",
	Short: "A web crawler daemon",
	Long:  `Crawld is a web crawler daemon for collecting and processing web content.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Crawld! Use --help for usage information.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
