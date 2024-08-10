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

type Task struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

const getTask = "/getTask/{taskId}"
const addTask = "/create"

func main() {
	repo := repositories.NewTodoRepository()
	service := services.NewTodoService(repo)
	handlers := handlers.NewTodoHandler(service)

	r := mux.NewRouter()
	r.HandleFunc(getTask, handlers.GetTask).Methods("GET")
	r.HandleFunc(addTask, handlers.CreateTask).Methods("POST")

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
