package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"memo/internal/models"
)

func setupTestHandler(t *testing.T) (*TodoHandler, func()) {
	// Create test database
	dbFile := "test_handler_todos.db"
	store, err := models.NewTodoStore(dbFile)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test templates
	templates := template.Must(template.New("test").Parse(`
		{{define "index.html"}}
		<html>
		<body>
			<div id="todos">
				{{range .Todos}}
					<div>{{.Text}} - {{.Completed}}</div>
				{{end}}
			</div>
		</body>
		</html>
		{{end}}

		{{define "todos.html"}}
		{{range .Todos}}
			<div>{{.Text}} - {{.Completed}}</div>
		{{end}}
		{{end}}
	`))

	handler := NewTodoHandler(store, templates)

	cleanup := func() {
		store.Close()
		os.Remove(dbFile)
	}

	return handler, cleanup
}

func TestIndexHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("renders index page with no todos", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.IndexHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "<html>") {
			t.Error("Expected HTML response")
		}
	})

	t.Run("renders index page with todos", func(t *testing.T) {
		// Add a test todo
		handler.store.AddTodo("Test todo")

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.IndexHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Test todo") {
			t.Error("Expected todo text in response")
		}
	})
}

func TestTodosHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("renders todos partial", func(t *testing.T) {
		handler.store.AddTodo("Test todo")

		req := httptest.NewRequest("GET", "/todos", nil)
		w := httptest.NewRecorder()

		handler.TodosHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Test todo") {
			t.Error("Expected todo text in response")
		}
	})
}

func TestAddTodoHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("adds todo successfully", func(t *testing.T) {
		form := url.Values{}
		form.Add("text", "New todo")

		req := httptest.NewRequest("POST", "/todos/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.AddTodoHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify todo was added
		todos, _ := handler.store.GetTodos()
		if len(todos) != 1 {
			t.Errorf("Expected 1 todo, got %d", len(todos))
		}
		if todos[0].Text != "New todo" {
			t.Errorf("Expected text 'New todo', got %q", todos[0].Text)
		}
	})

	t.Run("rejects empty todo text", func(t *testing.T) {
		form := url.Values{}
		form.Add("text", "")

		req := httptest.NewRequest("POST", "/todos/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.AddTodoHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("rejects non-POST requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/add", nil)
		w := httptest.NewRecorder()

		handler.AddTodoHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})

	t.Run("rejects text too long", func(t *testing.T) {
		form := url.Values{}
		form.Add("text", strings.Repeat("a", 501))

		req := httptest.NewRequest("POST", "/todos/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.AddTodoHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestToggleTodoHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("toggles existing todo", func(t *testing.T) {
		todo, _ := handler.store.AddTodo("Test todo")

		req := httptest.NewRequest("PUT", "/todos/toggle/"+strconv.Itoa(todo.ID), nil)
		w := httptest.NewRecorder()

		handler.ToggleTodoHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify todo was toggled
		todos, _ := handler.store.GetTodos()
		if len(todos) != 1 {
			t.Fatal("Expected 1 todo")
		}
		if !todos[0].Completed {
			t.Error("Expected todo to be completed")
		}
	})

	t.Run("returns 404 for non-existent todo", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/todos/toggle/99999", nil)
		w := httptest.NewRecorder()

		handler.ToggleTodoHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns 400 for invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/todos/toggle/invalid", nil)
		w := httptest.NewRecorder()

		handler.ToggleTodoHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("rejects non-PUT requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/toggle/1", nil)
		w := httptest.NewRecorder()

		handler.ToggleTodoHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestDeleteTodoHandler(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("deletes existing todo", func(t *testing.T) {
		todo, _ := handler.store.AddTodo("Test todo")

		req := httptest.NewRequest("DELETE", "/todos/delete/"+strconv.Itoa(todo.ID), nil)
		w := httptest.NewRecorder()

		handler.DeleteTodoHandler(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		// Verify todo was deleted
		todos, _ := handler.store.GetTodos()
		if len(todos) != 0 {
			t.Errorf("Expected 0 todos, got %d", len(todos))
		}
	})

	t.Run("returns 404 for non-existent todo", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/todos/delete/99999", nil)
		w := httptest.NewRecorder()

		handler.DeleteTodoHandler(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns 400 for invalid ID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/todos/delete/invalid", nil)
		w := httptest.NewRecorder()

		handler.DeleteTodoHandler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("rejects non-DELETE requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/todos/delete/1", nil)
		w := httptest.NewRecorder()

		handler.DeleteTodoHandler(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}