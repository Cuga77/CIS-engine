package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"

	rootCmd = &cobra.Command{
		Use:   "cis-cli",
		Short: "A CLI for interacting with the CIS search engine",
		Long:  `cis-cli is a command-line tool for searching documents and managing the CIS search engine.`,
	}
	apiBaseURL string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while executing CLI: '%s'", err)
		os.Exit(1)
	}
}

func init() {
	defaultURL := os.Getenv("CIS_API_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:8081"
	}
	rootCmd.PersistentFlags().StringVarP(&apiBaseURL, "api-url", "a", defaultURL, "Base URL for the search engine API (can be set via $CIS_API_URL)")
}