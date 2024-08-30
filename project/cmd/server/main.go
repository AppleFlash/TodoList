package main

import (
	"TaskManager/project/internal/handlers"
	"TaskManager/project/internal/repositories"
	"TaskManager/project/internal/services"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const getTask = "/getTask/{taskId}"
const addTask = "/create"
const register = "/register"
const login = "/login"

func main() {
	repo := repositories.NewTodoRepository()
	service := services.NewTodoService(repo)
	authRepo := repositories.NewAuthRepository()
	authService := services.NewAuthService(authRepo)
	handlers := handlers.NewTodoHandler(service, authService)

	r := mux.NewRouter()

	// r.HandleFunc(getTask, handlers.GetTask).Methods("GET")
	r.HandleFunc(addTask, handlers.CreateTask).Methods("POST")
	r.HandleFunc(register, handlers.Register).Methods("POST")
	r.HandleFunc(login, handlers.Login).Methods("POST")
	r.Use()

	createRoute(r, getTask, "GET", handlers.ProtectedMiddleware, handlers.GetTask)

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createRoute(
	r *mux.Router,
	endpoint string,
	method string,
	middleware mux.MiddlewareFunc,
	handler func(w http.ResponseWriter, r *http.Request),
) {
	protected := r.PathPrefix(endpoint).Subrouter()
	protected.Use(middleware)
	protected.HandleFunc("", handler).Methods(method)
}
