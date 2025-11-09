package cmd

import (
	"sync"
	"time"
)

var mu sync.Mutex

type Job struct {
	ID         int64     `json:"id"`
	Command    string    `json:"command"`
	Status     string    `json:"status"`
	Attempts   int       `json:"attempts"`
	MaxRetries int       `json:"max_retries"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Config struct {
	MaxRetries  int `json:"max_retries"`
	BackoffBase int `json:"backoff_base"`
}

var (
	jobQueue        []Job
	completedJobs   []Job
	deadLetterQueue []Job
	activeWorkers   int
	config          = Config{
		MaxRetries:  3,
		BackoffBase: 2,
	}
)

const (
	activeJobsFile    = "active_jobs.json"
	completedJobsFile = "completed_jobs.json"
	dlqJobsFile       = "dlq_jobs.json"
)
