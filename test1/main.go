package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Task представляет задачу в ToDo списке
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskService управляет задачами в памяти
type TaskService struct {
	tasks  map[int]*Task
	nextID int
	mutex  sync.RWMutex
}

// NewTaskService создает новый сервис задач
func NewTaskService() *TaskService {
	return &TaskService{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

// CreateTask создает новую задачу
func (ts *TaskService) CreateTask(title, description string) *Task {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	task := &Task{
		ID:          ts.nextID,
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ts.tasks[ts.nextID] = task
	ts.nextID++

	return task
}

// GetTask возвращает задачу по ID
func (ts *TaskService) GetTask(id int) (*Task, error) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	task, exists := ts.tasks[id]
	if !exists {
		return nil, fmt.Errorf("задача с ID %d не найдена", id)
	}

	return task, nil
}

// GetAllTasks возвращает все задачи
func (ts *TaskService) GetAllTasks() []*Task {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	tasks := make([]*Task, 0, len(ts.tasks))
	for _, task := range ts.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// UpdateTask обновляет существующую задачу
func (ts *TaskService) UpdateTask(id int, title, description string, completed bool) (*Task, error) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	task, exists := ts.tasks[id]
	if !exists {
		return nil, fmt.Errorf("задача с ID %d не найдена", id)
	}

	task.Title = title
	task.Description = description
	task.Completed = completed
	task.UpdatedAt = time.Now()

	return task, nil
}

// DeleteTask удаляет задачу по ID
func (ts *TaskService) DeleteTask(id int) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	_, exists := ts.tasks[id]
	if !exists {
		return fmt.Errorf("задача с ID %d не найдена", id)
	}

	delete(ts.tasks, id)
	return nil
}

// CreateTaskRequest представляет запрос на создание задачи
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// UpdateTaskRequest представляет запрос на обновление задачи
type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// TaskHandler обрабатывает HTTP запросы для задач
type TaskHandler struct {
	service *TaskService
}

// NewTaskHandler создает новый обработчик задач
func NewTaskHandler(service *TaskService) *TaskHandler {
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

func main() {
	// Создаем сервис и обработчик
	taskService := NewTaskService()
	taskHandler := NewTaskHandler(taskService)

	// Создаем роутер
	r := chi.NewRouter()

	// Добавляем middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Настраиваем CORS для тестирования с Postman
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Регистрируем маршруты
	r.Route("/tasks", func(r chi.Router) {
		r.Post("/", taskHandler.CreateTask)       // POST /tasks
		r.Get("/", taskHandler.GetTasks)          // GET /tasks
		r.Get("/{id}", taskHandler.GetTask)       // GET /tasks/{id}
		r.Put("/{id}", taskHandler.UpdateTask)    // PUT /tasks/{id}
		r.Delete("/{id}", taskHandler.DeleteTask) // DELETE /tasks/{id}
	})

	// Добавляем корневой маршрут для проверки
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":   "ToDo API работает!",
			"version":   "1.0.0",
			"endpoints": "POST /tasks, GET /tasks, GET /tasks/{id}, PUT /tasks/{id}, DELETE /tasks/{id}",
		})
	})

	// Запускаем сервер
	port := ":8080"
	fmt.Printf("🚀 Сервер запущен на http://localhost%s\n", port)
	fmt.Println("📋 Доступные эндпоинты:")
	fmt.Println("  POST   /tasks     - создать задачу")
	fmt.Println("  GET    /tasks     - получить все задачи")
	fmt.Println("  GET    /tasks/{id} - получить задачу по ID")
	fmt.Println("  PUT    /tasks/{id} - обновить задачу")
	fmt.Println("  DELETE /tasks/{id} - удалить задачу")
	fmt.Println("  GET    /           - информация об API")

	log.Fatal(http.ListenAndServe(port, r))
}
