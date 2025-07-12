package models

import (
	"database/sql"
	"log"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID        int       `json:"id"`
	Text      string    `json:"text"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type TodoStore struct {
	mu sync.RWMutex
	db *sql.DB
}

func NewTodoStore(dbPath string) (*TodoStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &TodoStore{db: db}
	if err := store.initDB(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (ts *TodoStore) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		text TEXT NOT NULL,
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := ts.db.Exec(query)
	if err != nil {
		log.Printf("Error creating todos table: %v", err)
		return err
	}

	return nil
}

func (ts *TodoStore) Close() error {
	return ts.db.Close()
}

func (ts *TodoStore) AddTodo(text string) Todo {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `INSERT INTO todos (text, completed, created_at) VALUES (?, ?, ?)`
	now := time.Now()
	
	result, err := ts.db.Exec(query, text, false, now)
	if err != nil {
		log.Printf("Error adding todo: %v", err)
		return Todo{}
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return Todo{}
	}

	return Todo{
		ID:        int(id),
		Text:      text,
		Completed: false,
		CreatedAt: now,
	}
}

func (ts *TodoStore) GetTodos() []Todo {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `SELECT id, text, completed, created_at FROM todos ORDER BY created_at DESC`
	rows, err := ts.db.Query(query)
	if err != nil {
		log.Printf("Error querying todos: %v", err)
		return []Todo{}
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Text, &todo.Completed, &todo.CreatedAt)
		if err != nil {
			log.Printf("Error scanning todo row: %v", err)
			continue
		}
		todos = append(todos, todo)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating todo rows: %v", err)
	}

	return todos
}

func (ts *TodoStore) ToggleTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `UPDATE todos SET completed = NOT completed WHERE id = ?`
	result, err := ts.db.Exec(query, id)
	if err != nil {
		log.Printf("Error toggling todo %d: %v", id, err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false
	}

	return rowsAffected > 0
}

func (ts *TodoStore) DeleteTodo(id int) bool {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `DELETE FROM todos WHERE id = ?`
	result, err := ts.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting todo %d: %v", id, err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false
	}

	return rowsAffected > 0
}