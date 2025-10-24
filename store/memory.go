package store

import (
	"errors"
	"log"
	"sync"
	md "taskflow/models"
)

type TaskStore struct {
	tasks map[string]*md.Task
	mu    sync.RWMutex
}

func NewTaskStore() *TaskStore {
	var ts TaskStore
	ts.tasks = make(map[string]*md.Task)
	return &ts
}

func (ts *TaskStore) Create(task *md.Task) {
	//lock the taskstore mutex
	ts.mu.Lock()

	//defer call to unlock the taskstore map
	defer ts.mu.Unlock()

	//Add task to the map
	ts.tasks[task.ID] = task
}

func (ts *TaskStore) Get(id string) (*md.Task, error) {
	//lock the taskstore mutex for reading
	ts.mu.RLock()

	//defer call to unlock the mutex for reading
	defer ts.mu.RUnlock()

	//get the task from map and if ok == false -> task not found
	task, ok := ts.tasks[id]
	if !ok {
		log.Println("ERROR: no task found using given taskid")
		return nil, errors.New("no task found using given task id")
	}

	//return the task
	return task, nil
}

func (ts *TaskStore) GetAll() []*md.Task {
	//lock the taskstore mutex for reading
	ts.mu.RLock()

	//defer call to unlock the mutex for reading
	defer ts.mu.RUnlock()

	//declare the slice of Task struct
	var task_list []*md.Task

	//iterate over map to append all the tasks in slice
	for _, task := range ts.tasks {
		task_list = append(task_list, task)
	}

	//return the slice
	return task_list
}

func (ts *TaskStore) Update(task *md.Task) error {
	//lock the taskstore mutex
	ts.mu.Lock()

	//defer call to unlock the taskstore map
	defer ts.mu.Unlock()

	//get the task id
	task_id := task.ID

	//check if task exists or not and return the error if it doesn't exists
	_, ok := ts.tasks[task_id]

	if !ok {
		log.Println("ERROR: no task found using given taskid")
		return errors.New("no task found using given task id")
	}

	//update the task in tasks
	ts.tasks[task_id] = task

	return nil
}
