package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Создаем сервис и обработчик
	taskService := NewTaskService()
	taskHandler := NewTaskHandler(taskService)

	// Настраиваем маршруты
	r := SetupRoutes(taskHandler)

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
