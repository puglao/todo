package router

import (
	"html/template"
	"net/http"

	"memo/internal/handlers"
	"memo/internal/models"
)

func SetupRoutes(store *models.TodoStore, templates *template.Template) *http.ServeMux {
	mux := http.NewServeMux()
	
	// Initialize handlers
	todoHandler := handlers.NewTodoHandler(store, templates)

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	mux.HandleFunc("/", todoHandler.IndexHandler)
	mux.HandleFunc("/todos", todoHandler.TodosHandler)
	mux.HandleFunc("/todos/add", todoHandler.AddTodoHandler)
	mux.HandleFunc("/todos/toggle/", todoHandler.ToggleTodoHandler)
	mux.HandleFunc("/todos/delete/", todoHandler.DeleteTodoHandler)

	return mux
}