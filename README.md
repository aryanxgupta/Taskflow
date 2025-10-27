# TaskFlow API

TaskFlow is a high-performance, concurrent job processing API written in Go. It's designed to accept tasks (that is processing a URL) asynchronously, execute them in a background worker pool, and store the results persistently in a PostgreSQL database or in memory storage.

This project is a deep dive into advanced Go concepts, including:

- **Concurrency**: Goroutines, Channels (`jobQueue`), `sync.WaitGroup` for graceful shutdown, and `sync.RWMutex` for thread-safe in-memory operations.
- **System Design**: Decoupling logic from storage using Interfaces (the `TaskStore` interface).
- **Database Integration**: Using `database/sql` and the `pgx` driver to connect to a real Postgres database.
- **Production-Ready Code**: Graceful shutdown, environment variable configuration, and robust error handling.

## Core Features

- **Asynchronous Job Processing**: The `POST /tasks` endpoint returns a `201 Accepted` response in milliseconds, while the actual work is processed in the background.
- **Concurrent Worker Pool**: A dispatcher manages a fixed-size pool of goroutine workers (default 5) that pull jobs from a buffered channel.
- **Persistent Storage**: Task status and results are stored in a PostgreSQL database, so no data is lost on restart.
- **Swappable Storage Layer**: Built on a `TaskStore` interface with two implementations:
  - `MemoryTaskStore`: A thread-safe, in-memory map (useful for testing).
  - `PostgresStore`: The primary, database-backed store (used by default).
- **Graceful Shutdown**: On Ctrl+C, the server stops accepting new requests, finishes in-flight work, and waits for all workers to complete their current jobs before exiting.
- **Context-Aware Workers**: Background HTTP-fetching jobs have a built-in timeout (5 seconds) using `context.WithTimeout`.

## Architecture

### API Handler (`api/handler.go`)
1. Receives an HTTP request (e.g., `POST /tasks`).
2. Validates the request body.
3. Creates a `Task` struct and saves it to the `TaskStore` (Postgres) with status: `PENDING`.
4. Submits the task to the Dispatcher.
5. Returns an immediate `201 Created` response.

### Dispatcher (`worker/dispatcher.go`)
- Manages a central `jobQueue` (a buffered Go channel).
- The `Submit()` method adds new tasks to this queue.
- On `Start()`, it launches a pool of worker goroutines.

### Worker (`worker/dispatcher.go`)
1. Each worker blocks and waits for a task to appear on the `jobQueue`.
2. Updates the task status to `RUNNING` in the `TaskStore`.
3. Performs the job (e.g., makes an HTTP call with a 5s timeout).
4. Updates the task status to `COMPLETED` or `FAILED` in the `TaskStore` with the result.

## 🚀 Getting Started

You can run the entire application locally using Go and a Docker-based Postgres instance.

### Prerequisites
- Go (1.21+)
- Docker Desktop
- A Git client

### 1. Clone the Repository
```bash
git clone https://github.com/aryanxgupta/Taskflow.git
cd taskflow
```

### 2. Install Dependencies
This project uses `godotenv` to load environment variables.
```bash
go mod tidy
```

### 3. Start the PostgreSQL Database
This command will start a Postgres container in the background, pre-configured with the correct user, password, and database.
```bash
docker run --name <container-name> -d \
  -e POSTGRES_USER=<user-name> \
  -e POSTGRES_PASSWORD=<password> \
  -e POSTGRES_DB=<db-name> \
  -p 5432:5432 \
  postgres
```

**To stop the container:** `docker stop <container-name>`
**To start it again:** `docker start <container-name>`

### 4. Create your Environment File
The server reads its configuration from a `.env` file. Create a file named `.env` in the root of the project:

**.env**
```
DATABASE_URL="postgres://<user-name>:<password>@localhost:5432/<db-name>?sslmode=disable"
```

### 5. Run the Application
```bash
go run ./cmd/taskflow/main.go
```

You should see the following output, confirming your database connection and server start:
```
2025/10/27 14:00:00 successfuly connected to PostgreSQL
2025/10/27 14:00:00 Dispatcher started with 5 workers
2025/10/27 14:00:00 Server is starting on http://localhost:8080
...
(Worker logs will appear here)
```

## API Endpoints

### POST /tasks
Creates a new task. The payload is the URL for the worker to fetch.

**Request:**
```bash
curl -X POST http://localhost:8080/tasks \
     -H "Content-Type: application/json" \
     -d '{"payload": "https://httpbin.org/delay/1"}'
```

**Success Response (201 Created):**
```json
{
    "id": "e522b00f-9508-4320-8be1-63dc6ecb9cbb",
    "payload": "https://httpbin.org/delay/1",
    "status": "PENDING",
    "result": null,
    "error": "",
    "created_at": "2025-10-27T14:02:00Z",
    "finished_at": "0001-01-01T00:00:00Z"
}
```

### GET /tasks
Retrieves a list of all tasks from the database.
```bash
curl http://localhost:8080/tasks
```

### GET /tasks/{id}
Retrieves the current status and result of a single task.
```bash
curl http://localhost:8080/tasks/e522b00f-9508-4320-8be1-63dc6ecb9cbb
```

**Success Response (200 OK):**
```json
{
    "id": "e522b00f-9508-4320-8be1-63dc6ecb9cbb",
    "payload": "https://httpbin.org/delay/1",
    "status": "COMPLETED",
    "result": {
        "message": "Request executed successfully...",
        "result": { ... },
        "status": "200 OK"
    },
    "error": "",
    "created_at": "2025-10-27T14:02:00Z",
    "finished_at": "2025-10-27T14:02:03Z"
}
```

## Running the Test Script

This project includes a simple shell script to stress-test the API and workers. It submits 6 tasks concurrently to test all success and failure cases.

### 1. Make the script executable (only need to do this once):
```bash
chmod +x run_tests.sh
```

### 2. Run the script:
```bash
./run_tests.sh
```

The script will run all 6 POST requests concurrently and will finish once all 6 requests have been sent.

### 3. Check the Results Manually
After a few seconds (to allow the workers to process the jobs), you can check the final results in your database by running GET requests:

**Get all tasks:**
```bash
curl http://localhost:8080/tasks
```

**Get a specific task (replace with an ID from your server logs):**
```bash
curl http://localhost:8080/tasks/YOUR_TASK_ID_HERE
```