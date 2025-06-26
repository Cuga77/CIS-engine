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
	Use:   "search [поисковый запрос]",
	Short: "Выполнить поиск документов",
	Long:  `Отправляет поисковый запрос к API и выводит найденные результаты.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Ошибка: неверный формат базового URL API: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/search"
		q := fullURL.Query()
		q.Set("q", query)
		fullURL.RawQuery = q.Encode()

		fmt.Printf("Отправка запроса на: %s\n", fullURL.String())

		resp, err := http.Get(fullURL.String())
		if err != nil {
			fmt.Printf("Ошибка при выполнении запроса к API: %v\n", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Ошибка при чтении ответа от API: %v\n", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("API вернуло ошибку (статус %d): %s\n", resp.StatusCode, string(body))
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
			fmt.Printf("Ошибка при парсинге JSON ответа от API: %v\n", err)
			return
		}

		fmt.Printf("\nРезультаты поиска по запросу \"%s\":\n", result.Query)
		if len(result.Results) == 0 {
			fmt.Println("Ничего не найдено.")
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
