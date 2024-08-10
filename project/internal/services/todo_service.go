package services

import (
	"TaskManager/project/internal/repositories"
	"TaskManager/project/pkg/models"
	"context"
)

type TodoService struct {
	repo repositories.TodoRepository
}

func NewTodoService(repo repositories.TodoRepository) *TodoService {
	return &TodoService{repo: repo}
}

func (s *TodoService) GetTask(ctx context.Context, id string) (models.Task, error) {
	return s.repo.GetTask(ctx, id)
}

func (s *TodoService) CreateTask(ctx context.Context, task models.Task) models.Task {
	s.repo.Create(ctx, task)
	return task
}
