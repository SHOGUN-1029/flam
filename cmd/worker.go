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
		//select {}
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		fmt.Println("Workers running... Press Ctrl+C to stop.")
		<-sigChan // waits here

		fmt.Println("\n Stopping workers...")
		stopWorkers()
	},
}

var workerStopCmd = &cobra.Command{
	Use:   "stop or ctrl+c",
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
		fmt.Println(" Invalid worker count.")
		return
	}

	stopChan = make(chan struct{})
	activeWorkers = count

	fmt.Printf(" Starting %d workers...\n", count)
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go workerLoop(i)
	}

	// Wait in background (so CLI doesn’t block forever)
	go func() {
		wg.Wait()
		fmt.Println(" All workers stopped.")
		activeWorkers = 0
	}()
}

// Worker main loop
func workerLoop(id int) {
	defer wg.Done()
	fmt.Printf(" Worker %d started.\n", id)

	for {
		select {
		case <-stopChan:
			fmt.Printf(" Worker %d stopping.\n", id)
			return
		default:
			job := getNextPendingJob()
			if job == nil {
				time.Sleep(1 * time.Second) // idle wait
				continue
			}
			executeJob(job, id)
		}
	}
}

func getNextPendingJob() *Job {
	mu.Lock()
	defer mu.Unlock()

	for i := range jobQueue {
		if jobQueue[i].Status == "pending" {
			jobQueue[i].Status = "processing"
			jobQueue[i].UpdatedAt = time.Now()
			return &jobQueue[i]
		}
	}
	return nil
}

func executeJob(job *Job, workerID int) {
	fmt.Printf(" Worker %d processing job %d (%s)\n", workerID, job.ID, job.Command)

	job.Attempts++
	job.UpdatedAt = time.Now()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", job.Command)
	} else {
		cmd = exec.Command("sh", "-c", job.Command)
	}

	output, err := cmd.CombinedOutput()
	fmt.Printf(" Worker %d output for job %d:\n%s\n", workerID, job.ID, string(output))

	mu.Lock()
	defer mu.Unlock()

	if err != nil {
		fmt.Printf(" Worker %d failed job %d: %v\n", workerID, job.ID, err)

		if job.Attempts >= config.MaxRetries {
			job.Status = "dead"
			deadLetterQueue = append(deadLetterQueue, *job)
			removeJobFromQueue(job.ID)
			fmt.Printf(" Job %d moved to DLQ after %d attempts\n", job.ID, job.Attempts)
		} else {
			// exponential backoff before requeue
			delay := int(math.Pow(float64(config.BackoffBase), float64(job.Attempts)))
			fmt.Printf(" Retrying job %d after %d seconds...\n", job.ID, delay)

			job.Status = "pending"
			job.UpdatedAt = time.Now().Add(time.Duration(delay) * time.Second)

			go func(retryJob Job, d int) {
				time.Sleep(time.Duration(d) * time.Second)
				mu.Lock()
				jobQueue = append(jobQueue, retryJob)
				mu.Unlock()
				_ = saveJobsToDiskLocked()
			}(*job, delay)
		}

		_ = saveJobsToDiskLocked()
		return
	}

	fmt.Printf(" Worker %d completed job %d\n", workerID, job.ID)
	job.Status = "completed"
	job.UpdatedAt = time.Now()

	completedJobs = append(completedJobs, *job)
	removeJobFromQueue(job.ID)
	_ = saveJobsToDiskLocked()
}

func removeJobFromQueue(jobID int64) {
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
	fmt.Println(" Workers stopped gracefully.")
}
