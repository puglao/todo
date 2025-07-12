package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"memo/internal/models"
)

type TodoHandler struct {
	store     *models.TodoStore
	templates *template.Template
}

func NewTodoHandler(store *models.TodoStore, templates *template.Template) *TodoHandler {
	return &TodoHandler{
		store:     store,
		templates: templates,
	}
}

func (h *TodoHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.GetTodos()
	if err != nil {
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []models.Todo
	}{
		Todos: todos,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TodoHandler) TodosHandler(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.GetTodos()
	if err != nil {
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []models.Todo
	}{
		Todos: todos,
	}

	if err := h.templates.ExecuteTemplate(w, "todos.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TodoHandler) AddTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text := r.FormValue("text")
	// Validation is now handled in the store layer
	_, err := h.store.AddTodo(text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	todos, err := h.store.GetTodos()
	if err != nil {
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []models.Todo
	}{
		Todos: todos,
	}

	if err := h.templates.ExecuteTemplate(w, "todos.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TodoHandler) ToggleTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/todos/toggle/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	err = h.store.ToggleTodo(id)
	if err != nil {
		if err.Error() == "todo not found" {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error updating todo", http.StatusInternalServerError)
		}
		return
	}

	todos, err := h.store.GetTodos()
	if err != nil {
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []models.Todo
	}{
		Todos: todos,
	}

	if err := h.templates.ExecuteTemplate(w, "todos.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TodoHandler) DeleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/todos/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	err = h.store.DeleteTodo(id)
	if err != nil {
		if err.Error() == "todo not found" {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting todo", http.StatusInternalServerError)
		}
		return
	}

	todos, err := h.store.GetTodos()
	if err != nil {
		http.Error(w, "Error retrieving todos", http.StatusInternalServerError)
		return
	}

	data := struct {
		Todos []models.Todo
	}{
		Todos: todos,
	}

	if err := h.templates.ExecuteTemplate(w, "todos.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}