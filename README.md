QueueCTL – CLI-Based Background Job Queue System
Tech Stack:

Language: Go (Golang)

Persistence: JSON file storage (no external database required)

Interface: Command Line Interface (Cobra-based)


Objective :

QueueCTL is a lightweight CLI-based background job queue system that allows users to enqueue shell commands, manage worker processes, handle retries with exponential backoff, and maintain a Dead Letter Queue (DLQ) for permanently failed jobs.
All data is persisted locally using JSON storage, ensuring durability across restarts.

Features :
 Enqueue and manage background jobs
Run multiple workers concurrently
Automatic retry with exponential backoff
Persistent JSON-based storage
Dead Letter Queue for failed jobs
Configurable retry and backoff parameters
Graceful shutdown handling
Clean and modular CLI design


Installation :
Option 1: Download Prebuilt Binary

Go to the Releases section.

Download the file based on the current os:
ex : queuectl-windows-amd64.exe
rename to queuectl (for easier usage)
Move it to a working directory  and open terminal in that directory.
manage permission if needed (for ex : chmod +x queuectl ;  allow queuectl in privacy and security settings in MacBook )

Option 2: Build from Source (optional)

If you prefer building yourself:

git clone https://github.com/SHOGUN-1029/flam
go build -o queuectl.exe main.go


Usage :

Once installed, you can run queuectl commands directly from Terminal:
example : Windows 
<img width="2559" height="1376" alt="image" src="https://github.com/user-attachments/assets/563bf76f-74ef-4e69-86d9-1d6aba7beae9" />
<img width="2559" height="939" alt="image" src="https://github.com/user-attachments/assets/d7e2539c-b1d0-48de-9d20-e2a6cd843a39" />
example : Mac
![WhatsApp Image 2025-11-09 at 13 59 14_5a0767ca](https://github.com/user-attachments/assets/dd13339a-eb47-479c-9de8-eb95efd57df1)

Core Commands :

| **Category**  | **Example Command**                     | **Description**              |
| ------------- | --------------------------------------- | ---------------------------- |
| **Enqueue**   | `./queuectl enqueue "echo Hello World"` | Add a job to the queue       |
| **Workers**   | `./queuectl worker start --count 3`     | Start worker processes       |
|               | `./queuectl worker stop` || ctrl +c     | Stop active workers          |
| **Status**    | `./queuectl status`                     | Display job & worker summary |
| **List Jobs** | `./queuectl list --state pending`       | List jobs by state           |
| **DLQ**       | `./queuectl dlq list`                   | View dead letter queue       |
|               | `./queuectl dlq retry <job_id>`         | Retry failed DLQ job         |
| **Config**    | `./queuectl config set max-retries 3`   | Set max retry count          |
|               | `./queuectl config show`                | View current configuration   |
| **Exit**      | `./queuectl exit`                       | Gracefully save and exit     |


Architecture Overview :

1. Job Lifecycle - 
| **State**    | **Meaning**                       |
| ------------ | --------------------------------- |
| `pending`    | Job waiting to be processed       |
| `processing` | Job currently being executed      |
| `completed`  | Job finished successfully         |
| `failed`     | Job failed but retryable          |
| `dead`       | Permanently failed (moved to DLQ) |

2. Retry & Backoff Logic - 
When a job fails, it retries automatically using:
delay = base ^ attempts
Example: base=2 → delays = 2s, 4s, 8s...
Once retries exceed max_retries, the job moves to DLQ.

3. Persistence
All queues (active, completed, dead) are stored as JSON and loaded back to slices when needed:
active_jobs.json
completed_jobs.json
dlq_jobs.json


Assumptions & Trade-Offs :

Designed for simplicity — uses JSON file persistence instead of databases.
Worker concurrency handled via Go routines (safe for single-machine use).
No external dependencies required — self-contained executable.
Platform-specific binary builds provided (no Docker setup required).

Evaluator Instructions :

Download the neccesary .exe (queuectl-windows-amd64.exe for windows ; queuectl-darwin-arm64 for Mac) binary from Releases.

Case 1 : Windows
Rename and open PowerShell in the same folder.

Run the following:

./queuectl enqueue "echo Hello"
./queuectl worker start
./queuectl list
./queuectl status
./queuectl dlq list

Verify that:
Jobs are processed successfully.
Restarting the app preserves job history.

Case 2 : Mac
Rename and open Terminal in the same folder.
enable permission :
chmod +x queuectl
go to privacy and setting and enable permissions under security for queuectl

Run the following:

./queuectl enqueue "echo Hello"
./queuectl worker start
./queuectl list
./queuectl status
./queuectl dlq list

Verify that:
Jobs are processed successfully.
Restarting the app preserves job history.

Directory Structure :

.
├── cmd/
│   ├── worker.go
│   ├── enqueue.go
│   ├── status.go
│   ├── list.go
│   ├── dlq.go
│   ├── config.go
│   ├── exit.go
│   └── storage.go
├── dist/
│   ├── queuectl-windows-amd64.exe
│   ├── queuectl-linux-amd64
│   ├── queuectl-darwin-arm64
│   └── ...
├── build-all.bat
├── main.go
├── go.mod
├── go.sum
├── active_jobs.json
├── completed_jobs.json
├── dlq_jobs.json
└── README.md


File Descriptions (with Example Commands):
 main.go
Entry point of the application. Initializes the CLI (Cobra root command).

cmd/worker.go
Handles worker management: starting, stopping, and processing jobs.
Implements retry, exponential backoff, and DLQ transitions.
Example:
./queuectl worker start --count 3
./queuectl worker stop

cmd/enqueue.go
Adds a new job to the queue also genertes unique id , time updated and made.
Jobs are stored persistently in active_jobs.json.
Example:
./queuectl enqueue "echo Hello"

cmd/status.go
Displays system status summary: pending, completed, failed, dead jobs, and active workers.
Example:
./queuectl status

cmd/list.go
Lists jobs filtered by their state or displays all if unspecified.
Example:
./queuectl list --state pending
./queuectl list

cmd/dlq.go
Manages the Dead Letter Queue (DLQ).
Lists dead jobs and requeues them for retry.
Example:
./queuectl dlq list
./queuectl dlq retry 1762582267013956200

cmd/config.go
Manages configuration parameters like max retries and backoff base.
Example:
./queuectl config show
./queuectl config set max-retries 4
./queuectl config set backoff-base 3

cmd/storage.go
Handles persistent storage for active, completed, and DLQ jobs (JSON-based).
Automatically called on startup and shutdown.
Example (internal):
# Automatically loads and saves job data

cmd/exit.go
Gracefully stops all workers, saves state, and shuts down QueueCTL.
Example:
./queuectl exit

build-all.bat
Builds platform-specific binaries for all major OS architectures.
Example:
.\build-all.bat

dist/
Contains precompiled binaries for multiple operating systems and architectures.

Checklist :
CLI Commands functional
Jobs persist after restart
Retry + backoff implemented
DLQ operational
Configurable parameters
Graceful shutdown
Clean documentation


Author :
Hemesh
Aspiring Computer Scientist | Backend Developer Intern Candidate
Built with Go 🐹 for performance, reliability, and simplicity.



