package models

import (
	"sync"
	"time"
)

type Todo struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type TodoStore struct {
	mu     sync.RWMutex
	todos  []Todo
	nextID int
}

func NewTodoStore() *TodoStore {
	return &TodoStore{
		todos:  []Todo{},
		nextID: 1,
	}
}

func (ts *TodoStore) AddTodo(text string) Todo {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	todo := Todo{
		ID:        ts.nextID,
		Text:      text,
		Completed: false,
		CreatedAt: time.Now(),
	}
	ts.todos = append(ts.todos, todo)
	ts.nextID++
	return todo
}

func (ts *TodoStore) GetTodos() []Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	todos := make([]Todo, len(ts.todos))
	copy(todos, ts.todos)
	return todos
}

func (ts *TodoStore) ToggleTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i := range ts.todos {
		if ts.todos[i].ID == id {
			ts.todos[i].Completed = !ts.todos[i].Completed
			return true
		}
	}
	return false
}

func (ts *TodoStore) DeleteTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos = append(ts.todos[:i], ts.todos[i+1:]...)
			return true
		}
	}
	return false
}