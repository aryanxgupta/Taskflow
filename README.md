TaskFlow - Asynchronous Go Job Processor (v1.0)

TaskFlow is a high-performance, asynchronous job processing API written in Go. This project demonstrates how to build a concurrent backend service that can receive tasks, process them in the background, and provide a way to track their status.

This v1.0 version is a complete, in-memory application that showcases several advanced Go concepts:

Concurrency: A robust worker pool using goroutines and buffered channels.

API Design: A clean HTTP API built with the chi router.

System Design: Decoupling the API (the "intake") from the workers (the "processing").

Graceful Shutdown: On Ctrl+C (SIGINT), the server stops accepting new requests, finishes all in-progress jobs, and then shuts down.

🏗️ Architecture Overview

The application is decoupled into several key components:

API (/internal/api): A chi router that exposes HTTP endpoints.

Store (/internal/store): A thread-safe, in-memory data store (using sync.RWMutex) for all tasks.

Dispatcher (/internal/worker): The "engine" of the app. It holds the buffered jobQueue channel and a sync.WaitGroup to manage the workers.

Workers: A pool of goroutines that listen on the jobQueue, pick up tasks, and execute them (simulated by a 5-second sleep).

Task Lifecycle:

Client -> POST /tasks with a JSON payload.

API Handler creates a Task struct and saves it to the MemoryStore with status: PENDING.

API Handler calls Dispatcher.Submit(&task).

API Handler immediately returns 201 Created with the task data.

A free Worker goroutine pulls the task from the jobQueue.

Worker updates the task status to RUNNING.

Worker performs the job (simulates work) and updates the task to COMPLETED.

