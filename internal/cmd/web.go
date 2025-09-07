package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/crush/internal/app"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/db"
	"github.com/charmbracelet/crush/internal/server"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface with full Crush backend integration",
	Long: `Start the Crush web interface server with complete backend integration.

The web interface provides a mobile-friendly, browser-based way to interact with Crush,
including Docker app building, chat functionality, and session management.
All features from the CLI are available through the web interface.`,
	Example: `
# Start web server on default port 8080
crush web

# Start on custom port  
crush web --port 3000

# Start with debug logging
crush web --debug`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		debug, _ := cmd.Flags().GetBool("debug")
		
		slog.Info("Initializing Crush web interface with backend integration", "port", port, "debug", debug)
		
		// Initialize configuration
		cfg, err := config.Init(".", "", debug)
		if err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		// Initialize database
		ctx := context.Background()
		dbConfig := cfg.Database
		if dbConfig == nil {
			// Default to SQLite
			dbConfig = &db.DatabaseConfig{
				Type:     "sqlite",
				Database: "crush.db",
				DataDir:  cfg.Options.DataDirectory,
			}
		}
		conn, err := db.Connect(ctx, dbConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer conn.Close()

		// Initialize the full Crush application
		crushApp, err := app.New(ctx, conn, cfg)
		if err != nil {
			return fmt.Errorf("failed to initialize Crush app: %w", err)
		}

		// Get services from the app
		agent := crushApp.CoderAgent
		sessions := crushApp.Sessions
		permissions := crushApp.Permissions

		slog.Info("Starting Crush web interface with full backend", "port", port)
		
		webServer := server.NewWebServer(port, agent, sessions, permissions)
		if err := webServer.Start(); err != nil {
			return fmt.Errorf("failed to start web server: %w", err)
		}
		
		return nil
	},
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
	webCmd.Flags().Bool("debug", false, "Enable debug logging")
	rootCmd.AddCommand(webCmd)
}