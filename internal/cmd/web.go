package cmd

import (
	"fmt"
	"log/slog"

	"github.com/charmbracelet/crush/internal/server"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web interface",
	Long: `Start the Crush web interface server.

The web interface provides a mobile-friendly, browser-based way to interact with Crush.
It includes an onboarding experience, documentation, and a chat interface.`,
	Example: `
# Start web server on default port 8080
crush web

# Start on custom port
crush web --port 3000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		
		slog.Info("Starting Crush web interface", "port", port)
		
		webServer := server.NewWebServer(port)
		if err := webServer.Start(); err != nil {
			return fmt.Errorf("failed to start web server: %w", err)
		}
		
		return nil
	},
}

func init() {
	webCmd.Flags().IntP("port", "p", 8080, "Port to run the web server on")
	rootCmd.AddCommand(webCmd)
}