package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/ryanvillarreal/krouter/pkg/config"
  "github.com/ryanvillarreal/krouter/pkg/router"
)

// Command execution through defined Run functions
var (
	verbose    bool
	configPath string
	rootCmd    = &cobra.Command{
		Use:   "krouter",
		Short: "krouter - A lightweight CoreDNS/CoreDHCP based router",
		Run: runRouter,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
  // idk if I want to do config as a persistent flag. the user could opt to run with the config in the local dir
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yml", "path to config file")
}

func runRouter(cmd *cobra.Command, args []string) {
	if verbose {
		log.Println("Verbose logging enabled")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// Create and start the router service
	svc,err := router.New(cfg)
	if err := svc.Start(); err != nil {
		log.Fatalf("Failed to start router: %v", err)
	}
  cfg.Display() // remove me after debugging or add to verbose output
	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	
	log.Printf("Received signal %v, shutting down...", sig)
	//svc.Stop()
}
