package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "queuectl",
	Short: "queuectl — A lightweight background job queue system",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := LoadJobsFromDisk(); err != nil {
			fmt.Printf(" Failed to load jobs from disk: %v\n", err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// All subcommands (enqueue, list, status, etc.) are added here.
}
