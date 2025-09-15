package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestTaskService_CreateTask(t *testing.T) {
	service := NewTaskService()

	task := service.CreateTask("Тестовая задача", "Описание тестовой задачи")

	if task.ID != 1 {
		t.Errorf("Ожидался ID = 1, получен %d", task.ID)
	}

	if task.Title != "Тестовая задача" {
		t.Errorf("Ожидался заголовок 'Тестовая задача', получен '%s'", task.Title)
	}

	if task.Description != "Описание тестовой задачи" {
		t.Errorf("Ожидалось описание 'Описание тестовой задачи', получено '%s'", task.Description)
	}

	if task.Completed != false {
		t.Errorf("Ожидалось Completed = false, получено %v", task.Completed)
	}

	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt не должно быть нулевым")
	}

	if task.UpdatedAt.IsZero() {
		t.Error("UpdatedAt не должно быть нулевым")
	}
}

func TestTaskService_GetTask(t *testing.T) {
	service := NewTaskService()

	// Создаем задачу
	createdTask := service.CreateTask("Тест", "Описание")

	// Получаем задачу
	retrievedTask, err := service.GetTask(createdTask.ID)
	if err != nil {
		t.Fatalf("Ошибка при получении задачи: %v", err)
	}

	if retrievedTask.ID != createdTask.ID {
		t.Errorf("ID не совпадает: ожидался %d, получен %d", createdTask.ID, retrievedTask.ID)
	}

	if retrievedTask.Title != createdTask.Title {
		t.Errorf("Заголовок не совпадает: ожидался '%s', получен '%s'", createdTask.Title, retrievedTask.Title)
	}
}

func TestTaskService_GetTask_NotFound(t *testing.T) {
	service := NewTaskService()

	_, err := service.GetTask(999)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей задачи")
	}

	expectedError := "задача с ID 999 не найдена"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestTaskService_GetAllTasks(t *testing.T) {
	service := NewTaskService()

	// Создаем несколько задач
	service.CreateTask("Задача 1", "Описание 1")
	service.CreateTask("Задача 2", "Описание 2")
	service.CreateTask("Задача 3", "Описание 3")

	tasks := service.GetAllTasks()

	if len(tasks) != 3 {
		t.Errorf("Ожидалось 3 задачи, получено %d", len(tasks))
	}

	// Проверяем, что все задачи имеют уникальные ID
	ids := make(map[int]bool)
	for _, task := range tasks {
		if ids[task.ID] {
			t.Errorf("Дублирующийся ID: %d", task.ID)
		}
		ids[task.ID] = true
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	service := NewTaskService()

	// Создаем задачу
	createdTask := service.CreateTask("Исходная задача", "Исходное описание")

	// Обновляем задачу
	updatedTask, err := service.UpdateTask(createdTask.ID, "Обновленная задача", "Обновленное описание", true)
	if err != nil {
		t.Fatalf("Ошибка при обновлении задачи: %v", err)
	}

	if updatedTask.Title != "Обновленная задача" {
		t.Errorf("Заголовок не обновлен: ожидался 'Обновленная задача', получен '%s'", updatedTask.Title)
	}

	if updatedTask.Description != "Обновленное описание" {
		t.Errorf("Описание не обновлено: ожидалось 'Обновленное описание', получено '%s'", updatedTask.Description)
	}

	if updatedTask.Completed != true {
		t.Errorf("Статус не обновлен: ожидалось true, получено %v", updatedTask.Completed)
	}

	if updatedTask.UpdatedAt.Before(createdTask.UpdatedAt) {
		t.Error("UpdatedAt должно быть больше исходного времени")
	}
}

func TestTaskService_UpdateTask_NotFound(t *testing.T) {
	service := NewTaskService()

	_, err := service.UpdateTask(999, "Новый заголовок", "Новое описание", true)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей задачи")
	}

	expectedError := "задача с ID 999 не найдена"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	service := NewTaskService()

	// Создаем задачу
	createdTask := service.CreateTask("Задача для удаления", "Описание")

	// Удаляем задачу
	err := service.DeleteTask(createdTask.ID)
	if err != nil {
		t.Fatalf("Ошибка при удалении задачи: %v", err)
	}

	// Проверяем, что задача удалена
	_, err = service.GetTask(createdTask.ID)
	if err == nil {
		t.Error("Задача должна быть удалена")
	}
}

func TestTaskService_DeleteTask_NotFound(t *testing.T) {
	service := NewTaskService()

	err := service.DeleteTask(999)
	if err == nil {
		t.Error("Ожидалась ошибка для несуществующей задачи")
	}

	expectedError := "задача с ID 999 не найдена"
	if err.Error() != expectedError {
		t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedError, err.Error())
	}
}

func TestTaskHandler_CreateTask(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	reqBody := CreateTaskRequest{
		Title:       "Тестовая задача",
		Description: "Описание тестовой задачи",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateTask(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusCreated, w.Code)
	}

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	if err != nil {
		t.Fatalf("Ошибка при парсинге ответа: %v", err)
	}

	if task.Title != reqBody.Title {
		t.Errorf("Заголовок не совпадает: ожидался '%s', получен '%s'", reqBody.Title, task.Title)
	}
}

func TestTaskHandler_CreateTask_EmptyTitle(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	reqBody := CreateTaskRequest{
		Title:       "",
		Description: "Описание без заголовка",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_CreateTask_InvalidJSON(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.CreateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_GetTasks(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	// Создаем несколько задач
	service.CreateTask("Задача 1", "Описание 1")
	service.CreateTask("Задача 2", "Описание 2")

	req := httptest.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	handler.GetTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
	}

	var tasks []Task
	err := json.Unmarshal(w.Body.Bytes(), &tasks)
	if err != nil {
		t.Fatalf("Ошибка при парсинге ответа: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Ожидалось 2 задачи, получено %d", len(tasks))
	}
}

func TestTaskHandler_GetTask(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	// Создаем задачу
	createdTask := service.CreateTask("Test Task", "Description")

	req := httptest.NewRequest("GET", "/tasks/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
	}

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	if err != nil {
		t.Fatalf("Ошибка при парсинге ответа: %v", err)
	}

	if task.ID != createdTask.ID {
		t.Errorf("ID не совпадает: ожидался %d, получен %d", createdTask.ID, task.ID)
	}
}

func TestTaskHandler_GetTask_NotFound(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	req := httptest.NewRequest("GET", "/tasks/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_GetTask_InvalidID(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	req := httptest.NewRequest("GET", "/tasks/invalid", nil)
	w := httptest.NewRecorder()
	handler.GetTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	// Создаем задачу
	service.CreateTask("Original Task", "Original Description")

	reqBody := UpdateTaskRequest{
		Title:       "Updated Task",
		Description: "Updated Description",
		Completed:   true,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusOK, w.Code)
	}

	var task Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	if err != nil {
		t.Fatalf("Ошибка при парсинге ответа: %v", err)
	}

	if task.Title != reqBody.Title {
		t.Errorf("Заголовок не обновлен: ожидался '%s', получен '%s'", reqBody.Title, task.Title)
	}

	if task.Completed != reqBody.Completed {
		t.Errorf("Статус не обновлен: ожидалось %v, получено %v", reqBody.Completed, task.Completed)
	}
}

func TestTaskHandler_UpdateTask_NotFound(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	reqBody := UpdateTaskRequest{
		Title:       "New Task",
		Description: "New Description",
		Completed:   true,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/tasks/999", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_UpdateTask_EmptyTitle(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	// Создаем задачу
	service.CreateTask("Original Task", "Original Description")

	reqBody := UpdateTaskRequest{
		Title:       "",
		Description: "Description without title",
		Completed:   false,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/tasks/1", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	// Создаем задачу
	service.CreateTask("Task to Delete", "Description")

	req := httptest.NewRequest("DELETE", "/tasks/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNoContent, w.Code)
	}

	// Проверяем, что задача действительно удалена
	req = httptest.NewRequest("GET", "/tasks/1", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w = httptest.NewRecorder()
	handler.GetTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Задача должна быть удалена, но получен статус %d", w.Code)
	}
}

func TestTaskHandler_DeleteTask_NotFound(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	req := httptest.NewRequest("DELETE", "/tasks/999", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusNotFound, w.Code)
	}
}

func TestTaskHandler_DeleteTask_InvalidID(t *testing.T) {
	service := NewTaskService()
	handler := NewTaskHandler(service)

	req := httptest.NewRequest("DELETE", "/tasks/invalid", nil)
	w := httptest.NewRecorder()
	handler.DeleteTask(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус %d, получен %d", http.StatusBadRequest, w.Code)
	}
}

// Тест для проверки конкурентности
func TestTaskService_Concurrency(t *testing.T) {
	service := NewTaskService()

	// Создаем несколько горутин для создания задач
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			service.CreateTask("Задача", "Описание")
			done <- true
		}(i)
	}

	// Ждем завершения всех горутин
	for i := 0; i < 10; i++ {
		<-done
	}

	tasks := service.GetAllTasks()
	if len(tasks) != 10 {
		t.Errorf("Ожидалось 10 задач, получено %d", len(tasks))
	}
}

// Тест для проверки времени создания и обновления
func TestTask_Timestamps(t *testing.T) {
	service := NewTaskService()

	beforeCreate := time.Now()
	task := service.CreateTask("Тест", "Описание")
	afterCreate := time.Now()

	if task.CreatedAt.Before(beforeCreate) || task.CreatedAt.After(afterCreate) {
		t.Error("CreatedAt должно быть между временем до и после создания")
	}

	if task.UpdatedAt.Before(beforeCreate) || task.UpdatedAt.After(afterCreate) {
		t.Error("UpdatedAt должно быть между временем до и после создания")
	}

	// Обновляем задачу
	time.Sleep(1 * time.Millisecond) // Небольшая задержка для различия во времени
	beforeUpdate := time.Now()
	service.UpdateTask(task.ID, "Обновлено", "Описание", true)
	afterUpdate := time.Now()

	updatedTask, _ := service.GetTask(task.ID)

	if updatedTask.UpdatedAt.Before(beforeUpdate) || updatedTask.UpdatedAt.After(afterUpdate) {
		t.Error("UpdatedAt должно быть обновлено при изменении задачи")
	}

	if updatedTask.CreatedAt != task.CreatedAt {
		t.Error("CreatedAt не должно изменяться при обновлении")
	}
}
