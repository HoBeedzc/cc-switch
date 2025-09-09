package web

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"time"

	"cc-switch/internal/handler"
)

//go:embed all:assets/*
var assets embed.FS

// Server represents the web server
type Server struct {
	handler handler.ConfigHandler
	server  *http.Server
	host    string
	port    int
}

// NewServer creates a new web server instance
func NewServer(configHandler handler.ConfigHandler, host string, port int) *Server {
	return &Server{
		handler: configHandler,
		host:    host,
		port:    port,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API routes
	api := &APIHandler{handler: s.handler}
	mux.HandleFunc("/api/profiles", api.HandleProfiles)
	mux.HandleFunc("/api/profiles/", api.HandleProfile)
	mux.HandleFunc("/api/current", api.HandleCurrent)
	mux.HandleFunc("/api/switch", api.HandleSwitch)
	mux.HandleFunc("/api/test", api.HandleTest)
	mux.HandleFunc("/api/templates", api.HandleTemplates)
	mux.HandleFunc("/api/templates/", api.HandleTemplateRoutes)
	mux.HandleFunc("/api/health", api.HandleHealth)
	mux.HandleFunc("/api/export", api.HandleExport)
	mux.HandleFunc("/api/import", api.HandleImport)

	// Static file server
	staticHandler := http.FileServer(http.FS(assets))
	mux.Handle("/assets/", staticHandler)

	// Main page
	mux.HandleFunc("/", s.handleIndex)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.host, s.port),
		Handler:      corsMiddleware(loggingMiddleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleIndex serves the main HTML page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>cc-switch Web Interface</title>
    <link rel="stylesheet" href="/assets/css/main.css">
</head>
<body>
    <div id="app">
        <header class="header">
            <div class="container">
                <h1>🔧 cc-switch</h1>
                <p class="subtitle">Claude Code Configuration Manager</p>
            </div>
        </header>
        <main class="main">
            <div class="container">
                <nav class="nav-tabs">
                    <button class="nav-tab active" data-section="profiles">Profiles</button>
                    <button class="nav-tab" data-section="templates">Templates</button>
                    <button class="nav-tab" data-section="settings">Settings</button>
                    <button class="nav-tab" data-section="test">API Test</button>
                </nav>
                
                <section id="profiles-section" class="section active">
                    <div class="section-header">
                        <h2>Configuration Profiles</h2>
                    </div>
                    <div class="section-content">
                        <div id="profiles-list">
                            <div class="loading">
                                <div class="spinner"></div>
                                Loading profiles...
                            </div>
                        </div>
                    </div>
                </section>
                
                <section id="templates-section" class="section">
                    <div class="section-header">
                        <h2>Template Management</h2>
                    </div>
                    <div class="section-content">
                        <div id="templates-list">
                            <div class="loading">
                                <div class="spinner"></div>
                                Loading templates...
                            </div>
                        </div>
                    </div>
                </section>
                
                <section id="settings-section" class="section">
                    <div class="section-header">
                        <h2>Settings</h2>
                    </div>
                    <div class="section-content" id="settings-content">
                        <div class="loading">
                            <div class="spinner"></div>
                            Loading settings...
                        </div>
                    </div>
                </section>
                
                <section id="test-section" class="section">
                    <div class="section-header">
                        <h2>API Connectivity Test</h2>
                    </div>
                    <div class="section-content" id="test-content">
                        <div class="loading">
                            <div class="spinner"></div>
                            Loading test interface...
                        </div>
                    </div>
                </section>
            </div>
        </main>
    </div>
    <script src="/assets/js/main.js"></script>
</body>
</html>`)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the response writer to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)
		fmt.Printf("%s %s %d %v\n", r.Method, r.URL.Path, wrapper.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
