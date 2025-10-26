package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"taskflow/models"
	"taskflow/store"
	"taskflow/worker"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TaskAPI struct {
	taskDispatcher *worker.Dispatcher
	taskStore      store.TaskStore
}

type CreateTaskRequest struct {
	Payload any `json:"payload"`
}

func NewTaskAPI(ts store.TaskStore, td *worker.Dispatcher) *TaskAPI {
	return &TaskAPI{
		taskDispatcher: td,
		taskStore:      ts,
	}
}

func (ta *TaskAPI) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR: invalid body: %v\n", err)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Payload == nil {
		log.Printf("ERROR: empty payload received")
		http.Error(w, "payload cannot be empty", http.StatusBadRequest)
		return
	}

	current_task := models.Task{
		ID:        uuid.New().String(),
		Payload:   req.Payload,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
	}

	ta.taskStore.Create(&current_task)
	ta.taskDispatcher.Submit(&current_task)

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(current_task)
}

func (ta *TaskAPI) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task_id := chi.URLParam(r, "id")

	current_task, err := ta.taskStore.Get(task_id)
	if err != nil {
		err_msg := fmt.Sprintf("invalid task id: %v", err.Error())
		log.Printf("ERROR: %s", err_msg)
		http.Error(w, err_msg, http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(current_task)
}

func (ta *TaskAPI) GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks := ta.taskStore.GetAll()

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}
