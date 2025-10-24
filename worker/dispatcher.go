package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	md "taskflow/models"
	ts "taskflow/store"
	"time"
)

const (
	DefaultMaxWorkers   = 5
	DefaultMaxQueueSize = 100
)

type Dispatcher struct {
	taskQueue chan *md.Task
	taskStore *ts.TaskStore
	workers   int
	wg        sync.WaitGroup
}

func NewDispatcher(store *ts.TaskStore, workers int) *Dispatcher {
	if workers == 0 {
		workers = DefaultMaxWorkers
	}

	return &Dispatcher{
		taskQueue: make(chan *md.Task, DefaultMaxQueueSize),
		taskStore: store,
		workers:   workers,
	}
}

func (ds *Dispatcher) workerFunc(i int) {
	log.Printf("Worker #%d started", i)
	defer ds.wg.Done()

	for task := range ds.taskQueue {
		log.Printf("Worker #%d picked Task #%s", i, task.ID)
		task.Status = md.StatusRunning
		ds.taskStore.Update(task)

		result, err := ds.processTask(task)
		if err != nil {
			log.Printf("ERROR: worker #%d failed task #%s: %v", i, task.ID, err)
			task.Status = md.StatusFailed
			task.Error = err.Error()

		} else {
			log.Printf("SUCCESS: worker #%d completed task #%s", i, task.ID)
			task.Status = md.StatusCompleted
			task.Result = result
		}

		task.FinishedAt = time.Now()
		ds.taskStore.Update(task)
	}

}

func (ds *Dispatcher) processTask(task *md.Task) ([]byte, error) {
	url, ok := task.Payload.(string)
	if !ok {
		log.Printf("ERROR: invalid url")
		return nil, fmt.Errorf("ERROR: invalid url: %s", task.Payload)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	new_request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("ERROR: unable to create a new request: %v", err)
		return nil, fmt.Errorf("ERROR: unable to create a new request: %v", err)
	}

	resp, err := http.DefaultClient.Do(new_request)
	if err != nil {
		log.Printf("ERROR: unable to execute the request: %v", err)
		return nil, fmt.Errorf("ERROR: unable to execute the request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("ERROR: http error: status: %s", resp.Status)
		return nil, fmt.Errorf("ERROR: http error: status: %s", resp.Status)
	}

	var p any
	err = json.NewDecoder(resp.Body).Decode(&p)

	if err != nil {
		log.Printf("ERROR: unable to decode the response body: %v", err)
		return nil, fmt.Errorf("ERROR: unable to decode the response body: %v", err)
	}

	result_data := map[string]interface{}{
		"message": fmt.Sprintf("Request executed successfully for URL %s", url),
		"result":  p,
		"status":  resp.Status,
	}

	success_bytes, err := json.Marshal(result_data)
	if err != nil {
		log.Printf("ERROR: unable to Marshal the response body: %v", err)
		return nil, fmt.Errorf("ERROR: unable to Marshal the response body: %v", err)
	}

	return success_bytes, nil
}

func (ds *Dispatcher) Start() {
	for i := 0; i < ds.workers; i++ {
		ds.wg.Add(1)
		go ds.workerFunc(i)
	}
	log.Printf("Dispatcher started with %d workers", ds.workers)
}

func (ds *Dispatcher) ShutDown() {
	log.Printf("Shutting down Dispatcher")

	close(ds.taskQueue)

	ds.wg.Wait()
	log.Println("All workers have finished. Dispatcher shutdown complete.")
}

func (ds *Dispatcher) Submit(task *md.Task) {
	ds.taskQueue <- task
}
