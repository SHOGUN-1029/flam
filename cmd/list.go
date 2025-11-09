package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

var stateFilter string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs in the queue by state",
	Run: func(cmd *cobra.Command, args []string) {
		filter := strings.ToLower(stateFilter)

		mu.Lock()

		allJobs := append([]Job{}, jobQueue...)
		allJobs = append(allJobs, completedJobs...)
		allJobs = append(allJobs, deadLetterQueue...)

		mu.Unlock()

		if len(allJobs) == 0 {
			fmt.Println(" No jobs available.")
			return
		}

		fmt.Printf("\n Listing jobs")
		if filter != "" {
			fmt.Printf(" with state: %s\n", filter)
		} else {
			fmt.Println(" (all states):")
		}

		fmt.Println("-----------------------------------------------------------------------")
		fmt.Printf("%-10s %-15s %-12s %-10s %-20s\n", "ID", "Command", "Status", "Attempts", "Updated At")
		fmt.Println("-----------------------------------------------------------------------")

		count := 0
		for _, job := range allJobs {
			if filter == "" || strings.ToLower(job.Status) == filter {
				// Truncate command for display
				cmdDisplay := job.Command
				if len(cmdDisplay) > 15 {
					cmdDisplay = cmdDisplay[:12] + "..."
				}

				fmt.Printf("%-10d %-15s %-12s %-10d %-20s\n",
					job.ID, cmdDisplay, job.Status, job.Attempts,
					job.UpdatedAt.Format(time.RFC822))
				count++
			}
		}

		if count == 0 {
			fmt.Println("  No jobs match the specified state.")
		}

		fmt.Println("-----------------------------------------------------------------------")
	},
}

func init() {
	listCmd.Flags().StringVar(&stateFilter, "state", "", "Filter jobs by state (pending, processing, completed, failed, dead)")
	rootCmd.AddCommand(listCmd)
}
