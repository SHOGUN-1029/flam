package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show summary of all job states and active workers",
	Run: func(cmd *cobra.Command, args []string) {
		showSystemStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func showSystemStatus() {
	// Lock before reading any shared slices
	mu.Lock()
	// Use defer to guarantee the lock is released when the function exits
	defer mu.Unlock()

	totalJobs := len(jobQueue) + len(completedJobs) + len(deadLetterQueue)

	if totalJobs == 0 {
		fmt.Println("📭 No jobs currently in system.")
		return
	}

	stateCount := map[string]int{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
		"dead":       0,
	}

	for _, job := range jobQueue {
		stateCount[job.Status]++
	}

	for _, job := range completedJobs {
		stateCount[job.Status]++
	}

	for _, job := range deadLetterQueue {
		stateCount[job.Status]++
	}

	fmt.Println(" QueueCTL System Status")
	fmt.Println("-------------------------")
	fmt.Println("Jobs Summary:")
	fmt.Printf("  Pending:     %d\n", stateCount["pending"])
	fmt.Printf("  Processing:  %d\n", stateCount["processing"])
	fmt.Printf("  Completed:   %d\n", stateCount["completed"])
	fmt.Printf("  Failed:      %d\n", stateCount["failed"])
	fmt.Printf("  Dead:        %d\n", stateCount["dead"])
	fmt.Printf("  Total Jobs:  %d\n", totalJobs)

	fmt.Println()
	fmt.Printf("Active Workers: %d\n", activeWorkers)
	fmt.Printf("DLQ Size:       %d\n", len(deadLetterQueue))

	fmt.Println()
	fmt.Println("Config:")
	fmt.Printf("  Max Retries:  %d\n", config.MaxRetries)
	fmt.Printf("  Backoff Base: %d\n", config.BackoffBase)
}
