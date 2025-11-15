package cmd

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	workerCount int
	stopChan    chan struct{}
	wg          sync.WaitGroup
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage background workers",
}

var workerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start one or more workers",
	Run: func(cmd *cobra.Command, args []string) {
		startWorkers(workerCount)

		// Wait for interrupt to stop workers gracefully
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		fmt.Println("Workers running... Press Ctrl+C to stop.")
		<-sigChan

		fmt.Println("\nStopping workers...")
		stopWorkers()
	},
}

var workerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all active workers gracefully",
	Run: func(cmd *cobra.Command, args []string) {
		stopWorkers()
	},
}

func init() {
	workerStartCmd.Flags().IntVar(&workerCount, "count", 1, "Number of workers to start")
	workerCmd.AddCommand(workerStartCmd)
	workerCmd.AddCommand(workerStopCmd)
	rootCmd.AddCommand(workerCmd)
}

func startWorkers(count int) {
	if count <= 0 {
		fmt.Println("Invalid worker count.")
		return
	}

	stopChan = make(chan struct{})
	activeWorkers = count

	fmt.Printf("Starting %d workers...\n", count)
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go workerLoop(i)
	}

	// background waiter to clear activeWorkers when all done
	go func() {
		wg.Wait()
		fmt.Println("All workers stopped.")
		activeWorkers = 0
	}()
}

// workerLoop picks jobs and executes them
func workerLoop(id int) {
	defer wg.Done()
	fmt.Printf("Worker %d started.\n", id)

	for {
		select {
		case <-stopChan:
			fmt.Printf("Worker %d stopping.\n", id)
			return
		default:
			index, job := getNextPendingJob()
			if job == nil {
				time.Sleep(1 * time.Second)
				continue
			}
			executeJob(index, job, id)
		}
	}
}

// getNextPendingJob returns index and pointer to next pending job (or -1, nil)
func getNextPendingJob() (int, *Job) {
	mu.Lock()
	defer mu.Unlock()

	for i := range jobQueue {
		if jobQueue[i].Status == "pending" {
			jobQueue[i].Status = "processing"
			jobQueue[i].UpdatedAt = time.Now()
			// Return index and pointer to element
			return i, &jobQueue[i]
		}
	}
	return -1, nil
}

// executeJob executes and handles retry / DLQ / completion
// index is the position in jobQueue for the job we picked (may become stale if removed elsewhere)
func executeJob(index int, job *Job, workerID int) {
	fmt.Printf("Worker %d processing job %d (%s)\n", workerID, job.ID, job.Command)

	// increment attempt before running (so attempt count reflects trials)
	job.Attempts++
	job.UpdatedAt = time.Now()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", job.Command)
	} else {
		cmd = exec.Command("sh", "-c", job.Command)
	}

	output, err := cmd.CombinedOutput()
	outStr := string(output)
	if outStr != "" {
		fmt.Printf("Worker %d output for job %d:\n%s\n", workerID, job.ID, outStr)
	} else {
		// still print empty output line for consistency
		fmt.Printf("Worker %d output for job %d:\n\n", workerID, job.ID)
	}

	// lock to update shared state
	mu.Lock()
	defer mu.Unlock()

	// locate the job in current queue by ID (index may be stale)
	realIdx := -1
	for i, j := range jobQueue {
		if j.ID == job.ID {
			realIdx = i
			break
		}
	}

	// Whether we find it or not, we will remove it from active queue now (if present).
	// This prevents duplicate processing / re-queuing the same active job.
	if realIdx != -1 {
		// take a snapshot of the job state (use copy) before removing
		currentJob := jobQueue[realIdx]
		// remove from jobQueue
		jobQueue = append(jobQueue[:realIdx], jobQueue[realIdx+1:]...)

		if err != nil {
			// failure branch
			fmt.Printf("Worker %d failed job %d: %v\n", workerID, currentJob.ID, err)

			if currentJob.Attempts >= currentJob.MaxRetries || currentJob.Attempts >= config.MaxRetries {
				// move to DLQ exactly once
				currentJob.Status = "dead"
				currentJob.UpdatedAt = time.Now()
				deadLetterQueue = append(deadLetterQueue, currentJob)
				_ = saveJobsToDiskLocked()
				fmt.Printf("Job %d moved to DLQ after %d attempts\n", currentJob.ID, currentJob.Attempts)
				return
			}

			// schedule retry with exponential backoff
			delay := int(math.Pow(float64(config.BackoffBase), float64(currentJob.Attempts)))
			if delay < 1 {
				delay = 1
			}
			fmt.Printf("Retrying job %d after %d seconds...\n", currentJob.ID, delay)

			// prepare retry job copy (status pending, keep attempts)
			retryJob := currentJob
			retryJob.Status = "pending"
			retryJob.UpdatedAt = time.Now().Add(time.Duration(delay) * time.Second)

			// schedule requeue in background (will acquire mu then save)
			go func(j Job, d int) {
				time.Sleep(time.Duration(d) * time.Second)
				mu.Lock()
				jobQueue = append(jobQueue, j)
				_ = saveJobsToDiskLocked()
				mu.Unlock()
			}(retryJob, delay)

			// persist current state: active job removed, DLQ unchanged
			_ = saveJobsToDiskLocked()
			return
		}

		// success branch
		currentJob.Status = "completed"
		currentJob.UpdatedAt = time.Now()
		completedJobs = append(completedJobs, currentJob)
		_ = saveJobsToDiskLocked()
		fmt.Printf("Worker %d completed job %d\n", workerID, currentJob.ID)
		return
	}

	// If we didn't find the job in jobQueue (rare), just handle safely:
	if err != nil {
		// treat as failure but do not duplicate DLQ; append to DLQ once
		job.Status = "dead"
		job.UpdatedAt = time.Now()
		deadLetterQueue = append(deadLetterQueue, *job)
		_ = saveJobsToDiskLocked()
		fmt.Printf("Worker %d failed job %d and moved to DLQ (not found in active queue)\n", workerID, job.ID)
		return
	}

	// success (not found in queue but command succeeded) -> append to completed
	job.Status = "completed"
	job.UpdatedAt = time.Now()
	completedJobs = append(completedJobs, *job)
	_ = saveJobsToDiskLocked()
	fmt.Printf("Worker %d completed job %d\n", workerID, job.ID)
}

// removeJobFromQueue removed a job by id (thread-safe)
func removeJobFromQueue(jobID int64) {
	mu.Lock()
	defer mu.Unlock()

	for i, job := range jobQueue {
		if job.ID == jobID {
			jobQueue = append(jobQueue[:i], jobQueue[i+1:]...)
			return
		}
	}
}

func stopWorkers() {
	if stopChan == nil {
		fmt.Println("No active workers running.")
		return
	}

	close(stopChan)
	wg.Wait()
	activeWorkers = 0
	fmt.Println("Workers stopped gracefully.")
}
