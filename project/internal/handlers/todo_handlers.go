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

func (h *TodoHandler) ProtectedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		if accessToken == "" {
			http.Error(w, "Access token is required", http.StatusBadRequest)
			return
		}

		if len(accessToken) > 7 && strings.ToUpper(accessToken[:7]) == "BEARER " {
			accessToken = accessToken[7:]
		}

		refreshToken := r.Header.Get("X-Refresh-Token")
		if refreshToken == "" {
			http.Error(w, "Refresh token is required", http.StatusBadRequest)
			return
		}

		accessData, error := h.authService.Validate(r.Context(), services.AccessData{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		})
		if error != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Add("Authorization", "Bearer "+accessData.AccessToken)
		w.Header().Set("X-Refresh-Token", accessData.RefreshToken)
		next.ServeHTTP(w, r)
	})
}
