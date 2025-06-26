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
	Short: "Добавить новый URL для сканирования",
	Long:  `Отправляет запрос на API, чтобы добавить новый URL в очередь на сканирование краулером.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		targetURL := args[0]

		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Ошибка: неверный формат базового URL API: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/crawl"

		requestBody, err := json.Marshal(map[string]string{"url": targetURL})
		if err != nil {
			fmt.Printf("Ошибка при создании JSON-запроса: %v\n", err)
			return
		}

		fmt.Printf("Отправка запроса на сканирование URL %s на эндпоинт: %s\n", targetURL, fullURL.String())

		resp, err := http.Post(fullURL.String(), "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			fmt.Printf("Ошибка при выполнении POST-запроса к API: %v\n", err)
			return
		}
		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Ошибка при чтении ответа от API: %v\n", err)
			return
		}

		fmt.Printf("Ответ сервера (статус %d): %s\n", resp.StatusCode, string(responseBody))
	},
}

func init() {
	rootCmd.AddCommand(crawlCmd)
}
