
# **QueueCTL – CLI-Based Background Job Queue System**

---

## **Tech Stack**

* **Language:** Go (Golang)
* **Persistence:** JSON file storage (no external database required)
* **Interface:** Command-Line Interface (Cobra-based)

---

## **Objective**

QueueCTL is a lightweight CLI-based background job queue system that allows users to enqueue shell commands, manage worker processes, handle retries with exponential backoff, and maintain a Dead Letter Queue (DLQ) for permanently failed jobs.
All data is persisted locally using JSON storage, ensuring durability across restarts.

---

## **Features**

* Enqueue and manage background jobs
* Run multiple workers concurrently
* Automatic retry with exponential backoff
* Persistent JSON-based storage
* Dead Letter Queue for failed jobs
* Configurable retry and backoff parameters
* Graceful shutdown handling
* Clean and modular CLI design

---

## **Installation**

### **Option 1: Download Prebuilt Binary**

1. Go to the **Releases** section.
2. Download the file based on your OS.

   * Example: `queuectl-windows-amd64.exe`
3. Rename to `queuectl` (for easier usage).
4. Move it to a working directory and open terminal in that directory.
5. Manage permissions if needed:

   * macOS/Linux:

     ```bash
     chmod +x queuectl
     ```

     Allow `queuectl` in **Privacy and Security settings** on macOS if required.

---

### **Option 2: Build from Source (optional)**

If you prefer building yourself:

```bash
git clone https://github.com/SHOGUN-1029/flam
go build -o queuectl.exe main.go
```

---

## **Usage**

Once installed, you can run `queuectl` commands directly from Terminal.

### **Example: Windows**

<img width="2559" height="1448" alt="image" src="https://github.com/user-attachments/assets/5e350c39-6855-417b-b79f-ebe8671f4a6a" />
<img width="2558" height="980" alt="image" src="https://github.com/user-attachments/assets/5a17362e-bdf8-4437-96df-1a10c2183082" />


### **Example: Mac**

![WhatsApp Image 2025-11-09 at 13 59 13_05bbf4e9](https://github.com/user-attachments/assets/d6a5e9ae-c628-4ca4-822f-ad8b129817f1)


---

## **Core Commands**

| **Category**  | **Example Command**                     | **Description**                |
| ------------- | --------------------------------------- | ------------------------------ |
| **Enqueue**   | `./queuectl enqueue "echo Hello World"` | Add a job to the queue         |
| **Workers**   | `./queuectl worker start --count 3`     | Start worker processes         |
|               | `./queuectl worker stop`                | Stop worker processes          |
| **Status**    | `./queuectl status`                     | Display job and worker summary |
| **List Jobs** | `./queuectl list --state pending`       | List jobs by state             |
| **DLQ**       | `./queuectl dlq list`                   | View Dead Letter Queue         |
|               | `./queuectl dlq retry <job_id>`         | Retry failed DLQ job           |
| **Config**    | `./queuectl config set max-retries 3`   | Set max retry count            |
|               | `./queuectl config show`                | View current configuration     |
| **Exit**      | `./queuectl exit`                       | Gracefully save and exit       |

---

## **Architecture Overview**

### **Job Lifecycle**

| **State**    | **Meaning**                       |
| ------------ | --------------------------------- |
| `pending`    | Job waiting to be processed       |
| `processing` | Job currently being executed      |
| `completed`  | Job finished successfully         |
| `failed`     | Job failed but retryable          |
| `dead`       | Permanently failed (moved to DLQ) |

---

### **Retry & Backoff Logic**

When a job fails, it retries automatically using:

```
delay = base ^ attempts
```

**Example:**
If base = 2 → delays = 2s, 4s, 8s...
Once retries exceed `max_retries`, the job moves to DLQ.

---

### **Persistence**

All queues (active, completed, dead) are stored as JSON and loaded back to slices when needed:

```
active_jobs.json
completed_jobs.json
dlq_jobs.json
```

---

## **Assumptions & Trade-Offs**

* Designed for simplicity — uses JSON file persistence instead of databases.
* Worker concurrency handled via Go routines (safe for single-machine use).
* No external dependencies — fully self-contained executable.
* Platform-specific binaries provided (no Docker setup required).

---

## **Evaluator Instructions**

### **Case 1: Windows**

1. Download the required `.exe` (`queuectl-windows-amd64.exe`) from **Releases**.
2. Rename and open **PowerShell** in the same folder.
3. Run:

   ```powershell
   ./queuectl enqueue "echo Hello"
   ./queuectl worker start
   ./queuectl list
   ./queuectl status
   ./queuectl dlq list
   ```
4. Verify:

   * Jobs are processed successfully.
   * Restarting the app preserves job history.

---

### **Case 2: Mac**

1. Download the binary (`queuectl-darwin-arm64`).
2. Rename and open **Terminal** in the same folder.
3. Enable permission:

   ```bash
   chmod +x queuectl
   ```

   Then go to **Privacy and Security** and allow execution.
4. Run:

   ```bash
   ./queuectl enqueue "echo Hello"
   ./queuectl worker start
   ./queuectl list
   ./queuectl status
   ./queuectl dlq list
   ```
5. Verify:

   * Jobs are processed successfully.
   * Restarting the app preserves job history.

---

## **Directory Structure**

```
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
```

---

## **File Descriptions (with Example Commands)**

### **main.go**

Entry point of the application. Initializes the CLI (Cobra root command).
**Example:**

```bash
./queuectl --help
```

---

### **cmd/worker.go**

Handles worker management — starting, stopping, and processing jobs.
Implements retry, exponential backoff, and DLQ transitions.
**Example:**

```bash
./queuectl worker start --count 3
./queuectl worker stop
```

---

### **cmd/enqueue.go**

Adds a new job to the queue, generates a unique ID, and timestamps creation and updates.
Jobs are persisted in `active_jobs.json`.
**Example:**

```bash
./queuectl enqueue "echo Hello"
```

---

### **cmd/status.go**

Displays system status: pending, completed, failed, dead jobs, and active workers.
**Example:**

```bash
./queuectl status
```

---

### **cmd/list.go**

Lists jobs filtered by state or displays all if unspecified.
**Example:**

```bash
./queuectl list --state pending
./queuectl list
```

---

### **cmd/dlq.go**

Manages the Dead Letter Queue (DLQ). Lists failed jobs and requeues them for retry.
**Example:**

```bash
./queuectl dlq list
./queuectl dlq retry 1762582267013956200
```

---

### **cmd/config.go**

Manages configuration parameters like max retries and backoff base.
**Example:**

```bash
./queuectl config show
./queuectl config set max-retries 4
./queuectl config set backoff-base 3
```

---

### **cmd/storage.go**

Handles persistent storage for active, completed, and DLQ jobs using JSON.
Automatically triggered on startup and shutdown.
**Example:** *(Internal only)*
Loads and saves job data automatically.

---

### **cmd/exit.go**

Gracefully stops all workers, saves state, and shuts down QueueCTL.
**Example:**

```bash
./queuectl exit
```

---

### **build-all.bat**

Builds platform-specific binaries for all major operating systems.
**Example:**

```bash
.\build-all.bat
```

---

### **dist/**

Contains precompiled binaries for multiple OS and architectures.
Used for distribution and evaluation.

---

## **Checklist**

* [x] CLI commands functional
* [x] Jobs persist after restart
* [x] Retry and backoff implemented
* [x] DLQ operational
* [x] Configurable parameters
* [x] Graceful shutdown
* [x] Clean documentation

---

## **Author**

**Hemesh**

Aspiring Computer Scientist | Backend Developer Intern Candidate
Built with Go for performance, reliability, and simplicity.
