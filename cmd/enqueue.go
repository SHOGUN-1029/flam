package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var enqueueCmd = &cobra.Command{
	Use:   "enqueue [command]",
	Short: "Add a job to the in-memory queue",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		command := args[0]
		job := Job{
			ID:         time.Now().UnixNano(),
			Command:    command,
			Status:     "pending",
			Attempts:   0,
			MaxRetries: config.MaxRetries,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// --- LOCK ---

		mu.Lock()
		jobQueue = append(jobQueue, job)

		queueSize := len(jobQueue)

		// Save the new job list to disk inside the lock to ensure consistency
		_ = saveJobsToDiskLocked()

		mu.Unlock()

		fmt.Printf("Job queued in memory: %s (Queue size: %d)\n", command, queueSize)
	},
}

func init() {
	rootCmd.AddCommand(enqueueCmd)
}
