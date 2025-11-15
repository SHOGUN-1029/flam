package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var dlqCmd = &cobra.Command{
	Use:   "dlq",
	Short: "Manage Dead Letter Queue (DLQ)",
	Long:  "Inspect and manage failed jobs moved to the Dead Letter Queue.",
}

var dlqListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all jobs in the Dead Letter Queue",
	Run: func(cmd *cobra.Command, args []string) {
		if err := LoadJobsFromDisk(); err != nil {
			fmt.Println("Failed to load jobs from disk:", err)
		}

		mu.Lock()
		dlqCopy := make([]Job, len(deadLetterQueue))
		copy(dlqCopy, deadLetterQueue)
		mu.Unlock()

		if len(dlqCopy) == 0 {
			fmt.Println("Dead Letter Queue is empty.")
			return
		}

		fmt.Println("Dead Letter Queue Jobs:")
		for _, job := range dlqCopy {
			fmt.Printf("• ID: %d | Command: %s | Attempts: %d/%d | Last Updated: %s\n",
				job.ID, job.Command, job.Attempts, job.MaxRetries, job.UpdatedAt.Format(time.RFC822))
		}
	},
}

var dlqRetryCmd = &cobra.Command{
	Use:   "retry <job_id>",
	Short: "Retry a specific job from the Dead Letter Queue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		jobID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			fmt.Println("Invalid job ID.")
			return
		}

		if err := LoadJobsFromDisk(); err != nil {
			fmt.Println("Failed to load jobs from disk:", err)
		}

		
		mu.Lock()
		idx := -1
		for i, j := range deadLetterQueue {
			if j.ID == jobID {
				idx = i
				break
			}
		}

		if idx == -1 {
			mu.Unlock()
			fmt.Printf("Job ID %d not found in DLQ.\n", jobID)
			return
		}

		jobToRetry := deadLetterQueue[idx]
		
		deadLetterQueue = append(deadLetterQueue[:idx], deadLetterQueue[idx+1:]...)

		
		jobToRetry.Status = "pending"
		jobToRetry.Attempts = 0
		jobToRetry.UpdatedAt = time.Now()
		jobQueue = append(jobQueue, jobToRetry)

		
		if err := saveJobsToDiskLocked(); err != nil {
			
			mu.Unlock()
			fmt.Println("Failed to persist queues after retry:", err)
			return
		}
		mu.Unlock()

		fmt.Printf("Job ID %d requeued successfully for processing.\n", jobID)
	},
}

func init() {
	dlqCmd.AddCommand(dlqListCmd)
	dlqCmd.AddCommand(dlqRetryCmd)
	rootCmd.AddCommand(dlqCmd)
}
