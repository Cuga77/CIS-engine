package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var crawlCmd = &cobra.Command{
	Use:   "crawl [url]",
	Short: "Add a new URL to be crawled",
	Long:  `Sends a request to the API to add a new URL to the crawler's queue.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetURL := args[0]

		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Error: invalid API base URL format: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/crawl"

		requestBody, err := json.Marshal(map[string]string{"url": targetURL})
		if err != nil {
			fmt.Printf("Error creating JSON request: %v\n", err)
			return
		}

		fmt.Printf("Sending request to crawl URL %s to endpoint: %s\n", targetURL, fullURL.String())

		resp, err := http.Post(fullURL.String(), "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("Error making POST request to API: %v\n", err)
			return
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response from API: %v\n", err)
			return
		}

		fmt.Printf("Server response (status %d): %s\n", resp.StatusCode, string(responseBody))
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
}