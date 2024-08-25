package handlers

import (
	"TaskManager/project/internal/services"
	"TaskManager/project/pkg/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type TodoHandler struct {
	service     *services.TodoService
	authService services.AuthService
}

func NewTodoHandler(service *services.TodoService, authService services.AuthService) *TodoHandler {
	return &TodoHandler{service: service, authService: authService}
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
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdTask := h.service.CreateTask(r.Context(), task)
	json.NewEncoder(w).Encode(createdTask)
}

func (h *TodoHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentionals
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.authService.Register(r.Context(), creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *TodoHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentionals
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := h.authService.Login(r.Context(), creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Add("Authorization", "Bearer "+data.Access.AccessToken)
	w.Header().Set("X-Refresh-Token", data.Access.RefreshToken)
}

func (h *TodoHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	if len(token) > 7 && strings.ToUpper(token[:7]) == "Bearer " {
		token = token[7:]
	}

	data, error := h.authService.Refresh(r.Context(), token)
	if error != nil {
		http.Error(w, error.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Add("Authorization", "Bearer "+data.AccessToken)
	w.Header().Set("X-Refresh-Token", data.RefreshToken)
}
