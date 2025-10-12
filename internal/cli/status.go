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
	Short: "Получить статус поискового движка",
	Long:  `Отправляет запрос к API для получения статистической информации о системе, такой как количество проиндексированных страниц.`,
	Run: func(cmd *cobra.Command, args []string) {
		fullURL, err := url.Parse(apiBaseURL)
		if err != nil {
			fmt.Printf("Ошибка: неверный формат базового URL API: %v\n", err)
			return
		}
		fullURL.Path = "/api/v1/status"

		fmt.Printf("Запрос статуса с эндпоинта: %s\n", fullURL.String())

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
			PagesCount int64 `json:"pages_count"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("Ошибка при парсинге JSON ответа от API: %v\n", err)
			return
		}

		fmt.Println("\n--- Статус Системы ---")
		fmt.Printf("Всего страниц в индексе: %d\n", result.PagesCount)
		fmt.Println("----------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
