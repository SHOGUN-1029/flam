package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadJobsFromDisk() error {
	if err := loadFromFile(activeFile, &jobQueue); err != nil {
		return fmt.Errorf("failed to load active jobs: %w", err)
	}

	if err := loadFromFile(completedFile, &completedJobs); err != nil {
		return fmt.Errorf("failed to load completed jobs: %w", err)
	}

	// Load DLQ
	if err := loadFromFile(dlqFile, &deadLetterQueue); err != nil {
		return fmt.Errorf("failed to load DLQ: %w", err)
	}

	return nil
}

func loadFromFile(filename string, target *[]Job) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			*target = []Job{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, target)
}

func saveJobsToDiskLocked() error {
	save := func(filename string, data []Job) error {
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filename, jsonData, 0644)
	}

	if err := save(activeFile, jobQueue); err != nil {
		return err
	}
	if err := save(completedFile, completedJobs); err != nil {
		return err
	}
	if err := save(dlqFile, deadLetterQueue); err != nil {
		return err
	}
	return nil
}

func SaveJobsToDisk() error {
	mu.Lock()
	defer mu.Unlock()
	return saveJobsToDiskLocked()
}
