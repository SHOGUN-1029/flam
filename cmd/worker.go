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

	
	go func() {
		wg.Wait()
		fmt.Println("All workers stopped.")
		activeWorkers = 0
	}()
}


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


func getNextPendingJob() (int, *Job) {
	mu.Lock()
	defer mu.Unlock()

	for i := range jobQueue {
		if jobQueue[i].Status == "pending" {
			jobQueue[i].Status = "processing"
			jobQueue[i].UpdatedAt = time.Now()
			
			return i, &jobQueue[i]
		}
	}
	return -1, nil
}


func executeJob(index int, job *Job, workerID int) {
	fmt.Printf("Worker %d processing job %d (%s)\n", workerID, job.ID, job.Command)

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
		fmt.Printf("Worker %d output for job %d:\n\n", workerID, job.ID)
	}

	mu.Lock()
	defer mu.Unlock()


	realIdx := -1
	for i, j := range jobQueue {
		if j.ID == job.ID {
			realIdx = i
			break
		}
	}


	if realIdx != -1 {		
		currentJob := jobQueue[realIdx]
		jobQueue = append(jobQueue[:realIdx], jobQueue[realIdx+1:]...)

		if err != nil {
			fmt.Printf("Worker %d failed job %d: %v\n", workerID, currentJob.ID, err)

			if currentJob.Attempts >= currentJob.MaxRetries || currentJob.Attempts >= config.MaxRetries {
				
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

			retryJob := currentJob
			retryJob.Status = "pending"
			retryJob.UpdatedAt = time.Now().Add(time.Duration(delay) * time.Second)

			go func(j Job, d int) {
				time.Sleep(time.Duration(d) * time.Second)
				mu.Lock()
				jobQueue = append(jobQueue, j)
				_ = saveJobsToDiskLocked()
				mu.Unlock()
			}(retryJob, delay)

			_ = saveJobsToDiskLocked()
			return
		}

		currentJob.Status = "completed"
		currentJob.UpdatedAt = time.Now()
		completedJobs = append(completedJobs, currentJob)
		_ = saveJobsToDiskLocked()
		fmt.Printf("Worker %d completed job %d\n", workerID, currentJob.ID)
		return
	}

	if err != nil {
		job.Status = "dead"
		job.UpdatedAt = time.Now()
		deadLetterQueue = append(deadLetterQueue, *job)
		_ = saveJobsToDiskLocked()
		fmt.Printf("Worker %d failed job %d and moved to DLQ (not found in active queue)\n", workerID, job.ID)
		return
	}

	job.Status = "completed"
	job.UpdatedAt = time.Now()
	completedJobs = append(completedJobs, *job)
	_ = saveJobsToDiskLocked()
	fmt.Printf("Worker %d completed job %d\n", workerID, job.ID)
}

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
