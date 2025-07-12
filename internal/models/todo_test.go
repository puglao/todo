package models

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*TodoStore, func()) {
	// Create temporary database file
	dbFile := "test_todos.db"
	
	store, err := NewTodoStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		store.Close()
		os.Remove(dbFile)
	}

	return store, cleanup
}

func TestNewTodoStore(t *testing.T) {
	t.Run("creates store with default path", func(t *testing.T) {
		store, cleanup := setupTestDB(t)
		defer cleanup()

		if store == nil {
			t.Fatal("Expected store to be created")
		}
	})

	t.Run("uses environment variable for database path", func(t *testing.T) {
		// Set environment variable
		testPath := "env_test.db"
		os.Setenv("DB_PATH", testPath)
		defer os.Unsetenv("DB_PATH")
		defer os.Remove(testPath)

		store, err := NewTodoStore("")
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		if store == nil {
			t.Fatal("Expected store to be created")
		}
	})

	t.Run("handles invalid database path", func(t *testing.T) {
		_, err := NewTodoStore("/invalid/path/todos.db")
		if err == nil {
			t.Fatal("Expected error for invalid path")
		}
	})
}

func TestValidateTodoText(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		wantErr bool
	}{
		{"valid text", "Buy groceries", false},
		{"empty text", "", true},
		{"whitespace only", "   ", true},
		{"text too long", strings.Repeat("a", 501), true},
		{"max length text", strings.Repeat("a", 500), false},
		{"text with newlines", "Line 1\nLine 2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTodoText(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTodoText() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddTodo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("adds valid todo", func(t *testing.T) {
		text := "Test todo"
		todo, err := store.AddTodo(text)
		if err != nil {
			t.Fatalf("AddTodo() error = %v", err)
		}

		if todo.ID == 0 {
			t.Error("Expected todo ID to be set")
		}
		if todo.Text != text {
			t.Errorf("Expected text %q, got %q", text, todo.Text)
		}
		if todo.Completed {
			t.Error("Expected new todo to be not completed")
		}
		if todo.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
	})

	t.Run("rejects empty text", func(t *testing.T) {
		_, err := store.AddTodo("")
		if err == nil {
			t.Error("Expected error for empty text")
		}
	})

	t.Run("rejects text too long", func(t *testing.T) {
		_, err := store.AddTodo(strings.Repeat("a", 501))
		if err == nil {
			t.Error("Expected error for text too long")
		}
	})

	t.Run("trims whitespace", func(t *testing.T) {
		text := "  Test with spaces  "
		todo, err := store.AddTodo(text)
		if err != nil {
			t.Fatalf("AddTodo() error = %v", err)
		}

		if todo.Text != strings.TrimSpace(text) {
			t.Errorf("Expected text to be trimmed, got %q", todo.Text)
		}
	})
}

func TestGetTodos(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("returns empty list initially", func(t *testing.T) {
		todos, err := store.GetTodos()
		if err != nil {
			t.Fatalf("GetTodos() error = %v", err)
		}

		if len(todos) != 0 {
			t.Errorf("Expected empty list, got %d todos", len(todos))
		}
	})

	t.Run("returns todos in descending order", func(t *testing.T) {
		// Add todos with slight delay to ensure different timestamps
		todo1, _ := store.AddTodo("First todo")
		time.Sleep(1 * time.Millisecond)
		todo2, _ := store.AddTodo("Second todo")
		time.Sleep(1 * time.Millisecond)
		todo3, _ := store.AddTodo("Third todo")

		todos, err := store.GetTodos()
		if err != nil {
			t.Fatalf("GetTodos() error = %v", err)
		}

		if len(todos) != 3 {
			t.Errorf("Expected 3 todos, got %d", len(todos))
		}

		// Should be in reverse chronological order (newest first)
		if todos[0].ID != todo3.ID {
			t.Errorf("Expected first todo to be %d, got %d", todo3.ID, todos[0].ID)
		}
		if todos[1].ID != todo2.ID {
			t.Errorf("Expected second todo to be %d, got %d", todo2.ID, todos[1].ID)
		}
		if todos[2].ID != todo1.ID {
			t.Errorf("Expected third todo to be %d, got %d", todo1.ID, todos[2].ID)
		}
	})
}

func TestToggleTodo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("toggles existing todo", func(t *testing.T) {
		todo, _ := store.AddTodo("Test todo")

		err := store.ToggleTodo(todo.ID)
		if err != nil {
			t.Fatalf("ToggleTodo() error = %v", err)
		}

		// Verify the todo was toggled
		todos, _ := store.GetTodos()
		if len(todos) != 1 {
			t.Fatal("Expected 1 todo")
		}
		if !todos[0].Completed {
			t.Error("Expected todo to be completed")
		}

		// Toggle again
		err = store.ToggleTodo(todo.ID)
		if err != nil {
			t.Fatalf("ToggleTodo() error = %v", err)
		}

		todos, _ = store.GetTodos()
		if todos[0].Completed {
			t.Error("Expected todo to be not completed")
		}
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		err := store.ToggleTodo(99999)
		if err == nil {
			t.Error("Expected error for non-existent todo")
		}
		if err.Error() != "todo not found" {
			t.Errorf("Expected 'todo not found' error, got %q", err.Error())
		}
	})

	t.Run("returns error for invalid ID", func(t *testing.T) {
		err := store.ToggleTodo(0)
		if err == nil {
			t.Error("Expected error for invalid ID")
		}
		if err.Error() != "invalid todo ID" {
			t.Errorf("Expected 'invalid todo ID' error, got %q", err.Error())
		}
	})
}

func TestDeleteTodo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("deletes existing todo", func(t *testing.T) {
		todo, _ := store.AddTodo("Test todo")

		err := store.DeleteTodo(todo.ID)
		if err != nil {
			t.Fatalf("DeleteTodo() error = %v", err)
		}

		// Verify the todo was deleted
		todos, _ := store.GetTodos()
		if len(todos) != 0 {
			t.Errorf("Expected 0 todos, got %d", len(todos))
		}
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		err := store.DeleteTodo(99999)
		if err == nil {
			t.Error("Expected error for non-existent todo")
		}
		if err.Error() != "todo not found" {
			t.Errorf("Expected 'todo not found' error, got %q", err.Error())
		}
	})

	t.Run("returns error for invalid ID", func(t *testing.T) {
		err := store.DeleteTodo(-1)
		if err == nil {
			t.Error("Expected error for invalid ID")
		}
		if err.Error() != "invalid todo ID" {
			t.Errorf("Expected 'invalid todo ID' error, got %q", err.Error())
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	// Test concurrent access to ensure thread safety
	t.Run("concurrent add operations", func(t *testing.T) {
		const numGoroutines = 10
		const todosPerGoroutine = 5
		
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(prefix int) {
				for j := 0; j < todosPerGoroutine; j++ {
					text := fmt.Sprintf("Todo %d-%d", prefix, j)
					_, err := store.AddTodo(text)
					if err != nil {
						results <- err
						return
					}
				}
				results <- nil
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			if err := <-results; err != nil {
				t.Errorf("Concurrent add failed: %v", err)
			}
		}

		// Verify all todos were added
		todos, err := store.GetTodos()
		if err != nil {
			t.Fatalf("GetTodos() error = %v", err)
		}

		expectedTotal := numGoroutines * todosPerGoroutine
		if len(todos) != expectedTotal {
			t.Errorf("Expected %d todos, got %d", expectedTotal, len(todos))
		}
	})
}

func TestDatabaseConnectionPool(t *testing.T) {
	t.Run("sets connection pool parameters", func(t *testing.T) {
		// Set environment variables
		os.Setenv("DB_MAX_OPEN_CONNS", "20")
		os.Setenv("DB_MAX_IDLE_CONNS", "10")
		os.Setenv("DB_CONN_MAX_LIFETIME", "2h")
		defer func() {
			os.Unsetenv("DB_MAX_OPEN_CONNS")
			os.Unsetenv("DB_MAX_IDLE_CONNS")
			os.Unsetenv("DB_CONN_MAX_LIFETIME")
		}()

		store, cleanup := setupTestDB(t)
		defer cleanup()

		// We can't directly test the connection pool settings,
		// but we can verify the store was created successfully
		if store == nil {
			t.Fatal("Expected store to be created")
		}
	})
}

