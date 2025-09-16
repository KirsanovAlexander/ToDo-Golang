package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupRoutes настраивает маршруты для приложения
func SetupRoutes(taskHandler *TaskHandler) *chi.Mux {
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

	return r
}
