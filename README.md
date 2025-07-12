# Memo - Todo List Application

A simple, fast todo list web application built with Go and HTMX. Features real-time updates without JavaScript and a clean, responsive interface.

## Features

- ✅ Add new todos
- ✅ Toggle todo completion status
- ✅ Delete todos
- ✅ Real-time updates with HTMX (no page reloads)
- ✅ Responsive design
- ✅ In-memory storage
- ✅ Graceful server shutdown
- ✅ Docker support

## Tech Stack

- **Backend**: Go 1.24 (standard library only)
- **Frontend**: HTML templates with HTMX
- **Storage**: In-memory thread-safe store
- **Styling**: CSS
- **Deployment**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Go 1.24 or later
- Docker (optional)

### Local Development

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd memo
   ```

2. Run the application:
   ```bash
   /usr/local/go/bin/go run cmd/server/main.go
   ```

3. Open your browser to `http://localhost:8080`

### Using Docker

1. Build and run with Docker Compose:
   ```bash
   docker-compose up -d
   ```

2. Access the application at `http://localhost:8080`

3. View logs:
   ```bash
   docker-compose logs -f memo-app
   ```

4. Stop the application:
   ```bash
   docker-compose down
   ```

### Building

To build the binary:
```bash
/usr/local/go/bin/go build -o bin/memo cmd/server/main.go
./bin/memo
```

## Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
memo/
├── cmd/server/main.go      # Application entry point
├── internal/
│   ├── models/todo.go      # Todo struct and thread-safe store
│   ├── handlers/todo.go    # HTTP request handlers
│   └── router/router.go    # Route configuration
├── templates/              # HTML templates
└── static/                 # CSS and static assets
```

### Key Components

- **TodoStore**: Thread-safe in-memory storage with mutex protection
- **Handlers**: HTTP request handlers for CRUD operations
- **Templates**: Go templates with HTMX integration
- **Server**: HTTP server with graceful shutdown and configurable timeouts

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Main page with todo list |
| `GET` | `/todos` | Todo list partial (HTMX target) |
| `POST` | `/todos/add` | Add new todo |
| `PUT` | `/todos/toggle/{id}` | Toggle todo completion |
| `DELETE` | `/todos/delete/{id}` | Delete todo |

## Development

The application uses Go's standard library exclusively, making it lightweight and dependency-free. Key features:

- **Graceful Shutdown**: Handles SIGTERM/SIGINT with 30-second timeout
- **Server Timeouts**: 15s read/write, 60s idle timeouts
- **Thread Safety**: Mutex-protected todo store for concurrent access
- **Template Caching**: Templates parsed once at startup
- **Static Files**: Served efficiently with `http.FileServer`

## HTMX Integration

The frontend uses HTMX for dynamic updates:

- **Add Todo**: Form submission updates the todo list
- **Toggle Status**: Checkbox clicks toggle completion
- **Delete Todo**: Button clicks remove todos with confirmation
- **No JavaScript**: All interactions handled by HTMX attributes

## Deployment

### Docker

The application includes a multi-stage Dockerfile for optimized production builds:

- **Build Stage**: Uses `golang:1.24-alpine` to compile the binary
- **Runtime Stage**: Uses `alpine:latest` for minimal image size
- **Health Check**: Included in docker-compose for monitoring

### Environment Variables

- `PORT`: Server port (default: 8080)

## License

This project is open source and available under the [MIT License](LICENSE).