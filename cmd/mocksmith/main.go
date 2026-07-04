// MockSmith - HTTP Mock Server & API Simulator
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EdgarOrtegaRamirez/mocksmith/internal/logger"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/models"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/parser"
	"github.com/EdgarOrtegaRamirez/mocksmith/internal/server"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := buildRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	var verbose bool
	var jsonLog bool

	root := &cobra.Command{
		Use:   "mocksmith",
		Short: "HTTP Mock Server & API Simulator",
		Long: `MockSmith is a config-driven HTTP mock server for API development,
testing, and prototyping. Define routes in YAML or JSON, and MockSmith
serves realistic responses with matching, delays, and hot-reload.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	root.PersistentFlags().BoolVar(&jsonLog, "json-log", false, "Output logs as JSON")

	// serve command
	serveCmd := &cobra.Command{
		Use:   "serve [config-file]",
		Short: "Start the mock server",
		Long:  `Start the mock HTTP server with the given configuration file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := args[0]
			watch, _ := cmd.Flags().GetBool("watch")

			cfg, err := parser.Load(configPath)
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			logLevel := "info"
			if verbose {
				logLevel = "debug"
			}

			log := logger.New(logLevel, jsonLog)

			log.LogEvent(logger.LevelInfo, "MockSmith v%s", version)
			log.LogEvent(logger.LevelInfo, "Routes loaded: %d", len(cfg.Routes))

			// Handle graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			srv := server.New(cfg, log)

			if watch {
				if err := srv.StartWatcher(configPath); err != nil {
					log.LogEvent(logger.LevelWarn, "File watcher not available: %v", err)
				} else {
					log.LogEvent(logger.LevelInfo, "Watching %s for changes", configPath)
				}
			}

			// Start server in a goroutine
			errCh := make(chan error, 1)
			go func() {
				errCh <- srv.Start()
			}()

			// Wait for interrupt or error
			select {
			case <-sigCh:
				log.LogEvent(logger.LevelInfo, "Shutting down gracefully...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return srv.Stop(ctx)
			case err := <-errCh:
				return err
			}
		},
	}
	serveCmd.Flags().BoolP("watch", "w", false, "Watch config file for changes")
	serveCmd.Flags().IntP("port", "p", 0, "Override server port")
	root.AddCommand(serveCmd)

	// validate command
	validateCmd := &cobra.Command{
		Use:   "validate [config-file]",
		Short: "Validate a configuration file",
		Long:  `Parse and validate a mock server configuration file without starting the server.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := parser.Load(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Valid configuration\n")
			fmt.Printf("   Routes: %d\n", len(cfg.Routes))
			fmt.Printf("   Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
			if cfg.CORS != nil {
				fmt.Printf("   CORS: enabled\n")
			}

			// List routes
			fmt.Printf("\nRoutes:\n")
			for i, route := range cfg.Routes {
				name := route.Name
				if name == "" {
					name = "(unnamed)"
				}
				fmt.Printf("  %d. %s %s [%s]\n", i+1, route.Method, route.Path, name)
				for j, resp := range route.Responses {
					respName := resp.Name
					if respName == "" {
						respName = fmt.Sprintf("response-%d", j+1)
					}
					fmt.Printf("     └─ %s: %d\n", respName, resp.StatusCode)
				}
			}

			return nil
		},
	}
	root.AddCommand(validateCmd)

	// routes command
	routesCmd := &cobra.Command{
		Use:   "routes [config-file]",
		Short: "List configured routes",
		Long:  `Display all configured routes in a formatted table.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := parser.Load(args[0])
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			outputFormat, _ := cmd.Flags().GetBool("json")

			if outputFormat {
				return printRoutesJSON(cfg)
			}

			printRoutesTable(cfg)
			return nil
		},
	}
	routesCmd.Flags().Bool("json", false, "Output as JSON")
	root.AddCommand(routesCmd)

	// sample command
	sampleCmd := &cobra.Command{
		Use:   "sample",
		Short: "Generate a sample configuration file",
		Long:  `Output a sample YAML configuration file to stdout.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Print(parser.SampleConfig())
			return nil
		},
	}
	root.AddCommand(sampleCmd)

	// version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("mocksmith v%s\n", version)
		},
	}
	root.AddCommand(versionCmd)

	return root
}

func printRoutesTable(cfg *models.Config) {
	fmt.Printf("%-4s %-8s %-25s %-20s %s\n", "#", "Method", "Path", "Name", "Status Codes")
	fmt.Printf("%-4s %-8s %-25s %-20s %s\n", "──", "──────", "────", "────", "─────────────")
	for i, route := range cfg.Routes {
		name := route.Name
		if name == "" {
			name = "-"
		}
		statusCodes := ""
		for j, resp := range route.Responses {
			if j > 0 {
				statusCodes += ", "
			}
			statusCodes += fmt.Sprintf("%d", resp.StatusCode)
		}
		fmt.Printf("%-4d %-8s %-25s %-20s %s\n", i+1, route.Method, route.Path, name, statusCodes)
	}
}

func printRoutesJSON(cfg *models.Config) error {
	data, err := json.MarshalIndent(cfg.Routes, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
