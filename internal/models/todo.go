package models

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"
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
	// Use environment variable if dbPath is empty
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "todos.db" // default fallback
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	maxOpenConns := 10
	maxIdleConns := 5
	connMaxLifetime := time.Hour

	// Allow override via environment variables
	if env := os.Getenv("DB_MAX_OPEN_CONNS"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val > 0 {
			maxOpenConns = val
		}
	}
	if env := os.Getenv("DB_MAX_IDLE_CONNS"); env != "" {
		if val, err := strconv.Atoi(env); err == nil && val > 0 {
			maxIdleConns = val
		}
	}
	if env := os.Getenv("DB_CONN_MAX_LIFETIME"); env != "" {
		if val, err := time.ParseDuration(env); err == nil && val > 0 {
			connMaxLifetime = val
		}
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

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

// validateTodoText validates todo text input
func validateTodoText(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errors.New("todo text cannot be empty")
	}
	if len(text) > 500 {
		return errors.New("todo text cannot exceed 500 characters")
	}
	return nil
}

func (ts *TodoStore) AddTodo(text string) (Todo, error) {
	// Validate input
	if err := validateTodoText(text); err != nil {
		return Todo{}, err
	}

	text = strings.TrimSpace(text)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `INSERT INTO todos (text, completed, created_at) VALUES (?, ?, ?)`
	now := time.Now()
	
	result, err := ts.db.Exec(query, text, false, now)
	if err != nil {
		log.Printf("Error adding todo: %v", err)
		return Todo{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		return Todo{}, err
	}

	return Todo{
		ID:        int(id),
		Text:      text,
		Completed: false,
		CreatedAt: now,
	}, nil
}

func (ts *TodoStore) GetTodos() ([]Todo, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	query := `SELECT id, text, completed, created_at FROM todos ORDER BY created_at DESC`
	rows, err := ts.db.Query(query)
	if err != nil {
		log.Printf("Error querying todos: %v", err)
		return nil, err
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
		return nil, err
	}

	return todos, nil
}

func (ts *TodoStore) ToggleTodo(id int) error {
	if id <= 0 {
		return errors.New("invalid todo ID")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `UPDATE todos SET completed = NOT completed WHERE id = ?`
	result, err := ts.db.Exec(query, id)
	if err != nil {
		log.Printf("Error toggling todo %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		return errors.New("todo not found")
	}

	return nil
}

func (ts *TodoStore) DeleteTodo(id int) error {
	if id <= 0 {
		return errors.New("invalid todo ID")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	query := `DELETE FROM todos WHERE id = ?`
	result, err := ts.db.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting todo %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		return errors.New("todo not found")
	}

	return nil
}