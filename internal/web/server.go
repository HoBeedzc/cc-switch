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
    <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='.9em' font-size='90'>🔧</text></svg>">
</head>
<body>
    <div id="app">
        <header class="header">
            <div class="container">
                <h1>🔧 CC-SWITCH</h1>
                <p class="subtitle">CLAUDE CODE CONFIGURATION MANAGER v1.0.0</p>
            </div>
        </header>
        
        <!-- Pixel art decorative border -->
        <div style="height: 8px; background: repeating-linear-gradient(to right, var(--pixel-teal) 0px, var(--pixel-teal) 8px, var(--pixel-purple) 8px, var(--pixel-purple) 16px, var(--pixel-pink) 16px, var(--pixel-pink) 24px, var(--pixel-blue) 24px, var(--pixel-blue) 32px);"></div>
        
        <main class="main">
            <div class="container">
                <!-- System Status Bar -->
                <div style="background: var(--dark-bg); color: var(--text-white); padding: 0.75rem 1.5rem; margin-bottom: 2rem; font-family: 'Press Start 2P', monospace; font-size: 0.6rem; letter-spacing: 1px; box-shadow: var(--shadow);">
                    <span style="color: var(--pixel-green);">●</span> SYSTEM ONLINE 
                    <span style="margin-left: 2rem; color: var(--pixel-teal);">●</span> PROFILES READY
                    <span style="margin-left: 2rem; color: var(--pixel-yellow);">●</span> STANDBY
                    <span style="float: right;">2025.09.11 | BUILD.001</span>
                </div>
                
                <nav class="nav-tabs">
                    <button class="nav-tab active" data-section="profiles">
                        <span style="margin-right: 0.5rem;">📋</span>PROFILES
                    </button>
                    <button class="nav-tab" data-section="templates">
                        <span style="margin-right: 0.5rem;">📋</span>TEMPLATES
                    </button>
                    <button class="nav-tab" data-section="settings">
                        <span style="margin-right: 0.5rem;">⚙️</span>SETTINGS
                    </button>
                    <button class="nav-tab" data-section="test">
                        <span style="margin-right: 0.5rem;">🔍</span>API TEST
                    </button>
                </nav>
                
                <section id="profiles-section" class="section active">
                    <div class="section-header">
                        <h2>📋 Configuration Profiles</h2>
                    </div>
                    <div class="section-content">
                        <div id="profiles-list">
                            <div class="loading">
                                <div class="spinner"></div>
                                LOADING PROFILES...
                            </div>
                        </div>
                    </div>
                </section>
                
                <section id="templates-section" class="section">
                    <div class="section-header">
                        <h2>📋 Template Management</h2>
                    </div>
                    <div class="section-content">
                        <div id="templates-list">
                            <div class="loading">
                                <div class="spinner"></div>
                                LOADING TEMPLATES...
                            </div>
                        </div>
                    </div>
                </section>
                
                <section id="settings-section" class="section">
                    <div class="section-header">
                        <h2>⚙️ System Settings</h2>
                    </div>
                    <div class="section-content" id="settings-content">
                        <div class="loading">
                            <div class="spinner"></div>
                            LOADING SETTINGS...
                        </div>
                    </div>
                </section>
                
                <section id="test-section" class="section">
                    <div class="section-header">
                        <h2>🔍 API Connectivity Test</h2>
                    </div>
                    <div class="section-content" id="test-content">
                        <div class="loading">
                            <div class="spinner"></div>
                            INITIALIZING TEST INTERFACE...
                        </div>
                    </div>
                </section>
            </div>
        </main>
        
        <!-- Pixel art footer -->
        <footer style="background: var(--dark-bg); color: var(--text-white); padding: 1rem 0; margin-top: 4rem;">
            <div class="container" style="text-align: center;">
                <p style="font-family: 'Press Start 2P', monospace; font-size: 0.6rem; letter-spacing: 1px;">
                    CC-SWITCH PIXEL INTERFACE v1.0.0 | 
                    <span style="color: var(--pixel-orange);">ANTHROPIC</span> | 
                    <span style="color: var(--pixel-teal);">CLAUDE CODE</span>
                </p>
            </div>
        </footer>
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
