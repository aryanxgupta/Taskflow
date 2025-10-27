package store

import "taskflow/models"

type TaskStore interface {
	Create(task *models.Task) error
	Get(id string) (*models.Task, error)
	GetAll() []*models.Task
	Update(task *models.Task) error
}
