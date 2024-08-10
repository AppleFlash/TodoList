package repositories

import (
	"TaskManager/project/pkg/models"
	"context"
	"errors"
	"sync"
)

type TodoRepository interface {
	GetTask(ctx context.Context, id string) (models.Task, error)
	Create(ctx context.Context, task models.Task)
}

type inMemoryTodoRepository struct {
	mu    sync.Mutex
	tasks map[string]*models.Task
}

func NewTodoRepository() TodoRepository {
	return &inMemoryTodoRepository{
		tasks: map[string]*models.Task{
			"id-1": {Id: "id-1", Name: "Foo"},
		},
	}
}

func (r *inMemoryTodoRepository) GetTask(ctx context.Context, id string) (models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if task, exists := r.tasks[id]; exists {
		return *task, nil
	} else {
		return models.Task{}, errors.New("Not found")
	}
}

func (r *inMemoryTodoRepository) Create(ctx context.Context, task models.Task) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks[task.Id] = &task
}
