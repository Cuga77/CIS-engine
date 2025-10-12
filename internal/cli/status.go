package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of the search engine",
	Long:  `Sends a request to the API to get statistical information about the system, such as the number of indexed pages.`,
	Run: func(cmd *cobra.Command, args []string) {
		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Error: invalid API base URL format: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/status"

		fmt.Printf("Requesting status from endpoint: %s\n", fullURL.String())

		resp, err := http.Get(fullURL.String())
		if err != nil {
			fmt.Printf("Error making GET request to API: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response from API: %v\n", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API returned an error (status %d): %s\n", resp.StatusCode, string(body))
			return
		}

		var result struct {
			PagesCount int64 `json:"pages_count"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("Error parsing JSON response from API: %v\n", err)
			return
		}

		fmt.Println("\n--- System Status ---")
		fmt.Printf("Total pages in index: %d\n", result.PagesCount)
		fmt.Println("---------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}