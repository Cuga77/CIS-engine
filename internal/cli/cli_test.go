package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrawlCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintln(w, `{"message":"URL accepted for scanning"}`)
	}))
	defer server.Close()

	apiBaseURL = server.URL

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"crawl", "http://example.com"})
	Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Server response (status 202): {\"message\":\"URL accepted for scanning\"}")
}

func TestSearchCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"query": "test",
			"results": []map[string]string{
				{"url": "http://example.com", "title": "Example"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiBaseURL = server.URL

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"search", "test"})
	Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Search results for query \"test\":")
	require.Contains(t, output, "1. Example")
}

func TestStatusCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]int64{"pages_count": 10}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	apiBaseURL = server.URL

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"status"})
	Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "Total pages in index: 10")
}

func TestVersionCmd(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"version"})
	Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "cis-cli version: dev")
}