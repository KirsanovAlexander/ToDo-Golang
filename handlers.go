package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// TaskHandler обрабатывает HTTP запросы для задач
type TaskHandler struct {
	service TaskServiceInterface
}

// NewTaskHandler создает новый обработчик задач
func NewTaskHandler(service TaskServiceInterface) *TaskHandler {
	return &TaskHandler{service: service}
}

// CreateTask обрабатывает POST /tasks
func (th *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Поле 'title' обязательно", http.StatusBadRequest)
		return
	}

	task := th.service.CreateTask(req.Title, req.Description)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// GetTasks обрабатывает GET /tasks
func (th *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := th.service.GetAllTasks()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// GetTask обрабатывает GET /tasks/{id}
func (th *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID задачи", http.StatusBadRequest)
		return
	}

	task, err := th.service.GetTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// UpdateTask обрабатывает PUT /tasks/{id}
func (th *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID задачи", http.StatusBadRequest)
		return
	}

	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Неверный JSON", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Поле 'title' обязательно", http.StatusBadRequest)
		return
	}

	task, err := th.service.UpdateTask(id, req.Title, req.Description, req.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// DeleteTask обрабатывает DELETE /tasks/{id}
func (th *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID задачи", http.StatusBadRequest)
		return
	}

	err = th.service.DeleteTask(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
