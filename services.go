package main

import (
	"fmt"
	"sync"
	"time"
)

// TaskServiceInterface определяет интерфейс для работы с задачами
type TaskServiceInterface interface {
	CreateTask(title, description string) *Task
	GetTask(id int) (*Task, error)
	GetAllTasks() []*Task
	UpdateTask(id int, title, description string, completed bool) (*Task, error)
	DeleteTask(id int) error
}

// TaskService управляет задачами в памяти
type TaskService struct {
	tasks  map[int]*Task
	nextID int
	mutex  sync.RWMutex
}

// NewTaskService создает новый сервис задач
func NewTaskService() TaskServiceInterface {
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
