package store

import "taskflow/models"

type TaskStore interface {
	Create(task *models.Task)
	Get(id string) (*models.Task, error)
	GetAll() []*models.Task
	Update(task *models.Task) error
}
