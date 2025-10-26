package store

import (
	"errors"
	"log"
	"sync"
	md "taskflow/models"
)

type MemoryTaskStore struct {
	tasks map[string]*md.Task
	mu    sync.RWMutex
}

func NewMemoryTaskStore() TaskStore {
	var ts MemoryTaskStore
	ts.tasks = make(map[string]*md.Task)
	return &ts
}

func (ts *MemoryTaskStore) Create(task *md.Task) {
	ts.mu.Lock()

	defer ts.mu.Unlock()

	ts.tasks[task.ID] = task
}

func (ts *MemoryTaskStore) Get(id string) (*md.Task, error) {
	ts.mu.RLock()

	defer ts.mu.RUnlock()

	task, ok := ts.tasks[id]
	if !ok {
		log.Println("ERROR: no task found using given taskid")
		return nil, errors.New("no task found using given task id")
	}

	return task, nil
}

func (ts *MemoryTaskStore) GetAll() []*md.Task {
	ts.mu.RLock()

	defer ts.mu.RUnlock()

	var task_list []*md.Task

	for _, task := range ts.tasks {
		task_list = append(task_list, task)
	}

	return task_list
}

func (ts *MemoryTaskStore) Update(task *md.Task) error {
	ts.mu.Lock()

	defer ts.mu.Unlock()

	task_id := task.ID

	_, ok := ts.tasks[task_id]

	if !ok {
		log.Println("ERROR: no task found using given taskid")
		return errors.New("no task found using given task id")
	}

	ts.tasks[task_id] = task

	return nil
}
