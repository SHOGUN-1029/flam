package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	activeFile    = "active_jobs.json"
	completedFile = "completed_jobs.json"
	dlqFile       = "dlq.json"
)

var exitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Gracefully stop workers and persist queue state",
	Run: func(cmd *cobra.Command, args []string) {
		gracefulExit()
	},
}

func init() {
	rootCmd.AddCommand(exitCmd)
}

func gracefulExit() {
	fmt.Println(" Initiating graceful shutdown...")

	if activeWorkers > 0 {
		stopWorkers()
	}

	if err := SaveJobsToDisk(); err != nil {
		fmt.Println("Failed to persist jobs:", err)
	} else {
		fmt.Println("All queue data saved successfully.")
	}

	fmt.Println(" Exiting QueueCTL. Goodbye!")
	os.Exit(0)
}
