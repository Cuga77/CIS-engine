package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Perform a document search",
	Long:  `Sends a search query to the API and prints the results.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Error: invalid API base URL format: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/search"
		q := fullURL.Query()
		q.Set("q", query)
		fullURL.RawQuery = q.Encode()

		fmt.Printf("Sending request to: %s\n", fullURL.String())

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
			Query   string `json:"query"`
			Results []struct {
				URL   string `json:"url"`
				Title string `json:"title"`
			} `json:"results"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("Error parsing JSON response from API: %v\n", err)
			return
		}

		fmt.Printf("\nSearch results for query \"%s\":\n", result.Query)
		if len(result.Results) == 0 {
			fmt.Println("No results found.")
			return
		}
		for i, r := range result.Results {
			fmt.Printf("%d. %s\n   %s\n", i+1, r.Title, r.URL)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}