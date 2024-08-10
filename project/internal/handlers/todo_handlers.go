package handlers

import (
	"TaskManager/project/internal/services"
	"TaskManager/project/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	service *services.TodoService
}

func NewTodoHandler(service *services.TodoService) *TodoHandler {
	return &TodoHandler{service: service}
}

func (h *TodoHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId, exists := vars["taskId"]
	if !exists {
		http.Error(w, "taskId not provided", http.StatusBadRequest)
		return
	}

	fmt.Println("Start finding task at id: ", taskId)
	time.Sleep(time.Second)
	task, error := h.service.GetTask(r.Context(), taskId)
	if error != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TodoHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdTask := h.service.CreateTask(r.Context(), task)
	json.NewEncoder(w).Encode(createdTask)
}
