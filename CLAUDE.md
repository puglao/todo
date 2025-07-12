# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A todo-list web application built with Go and HTMX. The app provides real-time todo management using HTMX for dynamic updates without page reloads.

## Development Commands

### Local Development
- **Run the application**: `/usr/local/go/bin/go run cmd/server/main.go`
- **Build the application**: `/usr/local/go/bin/go build -o bin/memo cmd/server/main.go`
- **Run built binary**: `./bin/memo`

### Docker Commands
- **Build Docker image**: `docker build -t memo-app .`
- **Run Docker container**: `docker run -p 8080:8080 memo-app`
- **Run with docker-compose**: `docker-compose up -d`
- **Stop docker-compose**: `docker-compose down`
- **View logs**: `docker-compose logs -f memo-app`

The server runs on `http://localhost:8080`

**Go Version**: 1.24 (ARM64 architecture optimized for Apple Silicon devcontainers)

## Project Structure

```
memo/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── models/
│   │   └── todo.go          # Todo struct and TodoStore
│   ├── handlers/
│   │   └── todo.go          # HTTP request handlers
│   └── router/
│       └── router.go        # Route configuration
├── templates/               # HTML templates
│   ├── index.html          # Main page layout with HTMX
│   └── todos.html          # Todo list partial template
├── static/                 # Static assets
│   └── style.css           # Application styles
├── bin/                    # Built binaries (created during build)
├── Dockerfile              # Docker image configuration
├── docker-compose.yml      # Docker Compose configuration
├── .dockerignore           # Docker ignore patterns
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── .gitignore              # Git ignore patterns
├── .cursorignore           # Cursor editor ignore patterns
└── .env                    # Environment variables (empty)
```

## Architecture

The application uses:

- **Go 1.24 standard library** with improved HTTP server features
- **Graceful shutdown** with context-based timeout handling
- **Enhanced HTTP server** with configurable timeouts and proper mux routing
- **In-memory storage** with thread-safe TodoStore for data persistence
- **HTMX** for dynamic UI updates without JavaScript
- **Template-based rendering** for server-side HTML generation

### Key Components

- **`internal/models/todo.go`**: Todo struct and thread-safe TodoStore with CRUD operations
- **`internal/handlers/todo.go`**: HTTP request handlers for todo operations
- **`internal/router/router.go`**: Route configuration and setup
- **`cmd/server/main.go`**: Application entry point with server configuration
- **HTTP Server**: Go 1.24 server with graceful shutdown, configurable timeouts (15s read/write, 60s idle)
- **Signal handling**: Proper SIGTERM/SIGINT handling for clean shutdown
- Template system: Uses Go's `html/template` for server-side rendering
- HTMX integration: Enables real-time updates for add, toggle, and delete operations

### API Endpoints

- `GET /`: Main page with todo list
- `GET /todos`: Todo list partial (for HTMX updates)
- `POST /todos/add`: Add new todo
- `PUT /todos/toggle/{id}`: Toggle todo completion
- `DELETE /todos/delete/{id}`: Delete todo