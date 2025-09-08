package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cc-switch/internal/config"
	"cc-switch/internal/handler"
	"cc-switch/internal/web"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	webPort      int
	webHost      string
	webNoBrowser bool
	webQuiet     bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Launch web interface for cc-switch",
	Long: `Start a web server that provides a browser-based interface for managing Claude Code configurations.

The web interface allows you to:
- View all available configurations
- Switch between configurations
- Create, edit, and delete configurations
- Test API connectivity
- Export and import configurations

The server will be available at http://localhost:13501 (or custom host:port)
By default, your web browser will open automatically to the interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkClaudeConfig(); err != nil {
			return err
		}

		// Initialize config manager and handler
		cm, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("failed to initialize config manager: %w", err)
		}

		configHandler := handler.NewConfigHandler(cm)

		// Check if port is available
		if err := checkPortAvailable(webHost, webPort); err != nil {
			return fmt.Errorf("port %d is not available: %w", webPort, err)
		}

		// Create web server
		server := web.NewServer(configHandler, webHost, webPort)

		// Start server in goroutine
		serverErr := make(chan error, 1)
		go func() {
			if !webQuiet {
				color.Green("🚀 Starting cc-switch web interface...")
				fmt.Printf("📍 Server: http://%s:%d\n", webHost, webPort)
				fmt.Printf("💡 Press Ctrl+C to stop\n\n")
			}

			if err := server.Start(); err != nil {
				serverErr <- err
			}
		}()

		// Open browser automatically unless --no-browser is specified
		if !webNoBrowser {
			time.Sleep(500 * time.Millisecond) // Give server time to start
			go openBrowser(fmt.Sprintf("http://%s:%d", webHost, webPort))
		}

		// Setup graceful shutdown
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

		select {
		case err := <-serverErr:
			return fmt.Errorf("server failed to start: %w", err)
		case <-shutdown:
			if !webQuiet {
				fmt.Println("\n🛑 Shutting down server...")
			}

			// Create shutdown context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				return fmt.Errorf("server shutdown failed: %w", err)
			}

			if !webQuiet {
				color.Green("✅ Server stopped gracefully")
			}
			return nil
		}
	},
}

func init() {
	webCmd.Flags().IntVarP(&webPort, "port", "p", 13501, "Port to serve on")
	webCmd.Flags().StringVarP(&webHost, "host", "H", "localhost", "Host to bind to")
	webCmd.Flags().BoolVarP(&webNoBrowser, "no-browser", "n", false, "Don't open browser automatically")
	webCmd.Flags().BoolVarP(&webQuiet, "quiet", "q", false, "Suppress startup messages")
}

// checkPortAvailable checks if a port is available
func checkPortAvailable(host string, port int) error {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	return nil
}

// openBrowser attempts to open the URL in the default browser
func openBrowser(url string) {
	var cmd string
	var args []string

	switch {
	case fileExists("/usr/bin/xdg-open"): // Linux
		cmd = "xdg-open"
		args = []string{url}
	case fileExists("/usr/bin/open"): // macOS
		cmd = "open"
		args = []string{url}
	case fileExists("/c/Windows/System32/rundll32.exe"): // Windows
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		fmt.Printf("💻 Please open your browser and go to: %s\n", url)
		return
	}

	// Execute the command to open browser
	exec := &exec.Cmd{
		Path: cmd,
		Args: append([]string{cmd}, args...),
	}

	if err := exec.Start(); err != nil {
		fmt.Printf("💻 Please open your browser and go to: %s\n", url)
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
