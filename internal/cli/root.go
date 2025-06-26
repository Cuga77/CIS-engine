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
		Short: "CLI для взаимодействия с поисковым движком CIS",
		Long:  `cis-cli - это инструмент командной строки для поиска документов и управления поисковым движком CIS.`,
	}
	apiBaseURL string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка при выполнении CLI: '%s'", err)
		os.Exit(1)
	}
}

func init() {
	defaultURL := os.Getenv("CIS_API_URL")
	if defaultURL == "" {
		defaultURL = "http://51.250.38.170"
	}
	rootCmd.PersistentFlags().StringVarP(&apiBaseURL, "api-url", "a", defaultURL, "Базовый URL API поискового движка (можно задать через $CIS_API_URL)")
}
