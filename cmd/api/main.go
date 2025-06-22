package main

import (
	"context"
	"log"
	"os"

	"cis-engine/internal/api"
	"cis-engine/internal/search"
	"cis-engine/internal/storage/postgres"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используются переменные окружения системы")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Переменная окружения DATABASE_URL не установлена")
	}

	ctx := context.Background()
	db, err := postgres.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()
	log.Println("Успешное подключение к базе данных.")

	searchService := search.NewService(db)
	apiHandler := api.NewHandler(searchService)
	router := api.NewRouter(apiHandler)

	serverAddr := ":8080"
	log.Printf("Запуск API сервера на http://localhost%s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}
}
