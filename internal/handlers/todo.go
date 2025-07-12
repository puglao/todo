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
	todos := h.store.GetTodos()
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
	todos := h.store.GetTodos()
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
	if text == "" {
		http.Error(w, "Todo text is required", http.StatusBadRequest)
		return
	}

	h.store.AddTodo(text)

	todos := h.store.GetTodos()
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

	if !h.store.ToggleTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos := h.store.GetTodos()
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

	if !h.store.DeleteTodo(id) {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	todos := h.store.GetTodos()
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