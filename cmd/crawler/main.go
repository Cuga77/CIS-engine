package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cis-engine/internal/crawler"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := postgres.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()
	log.Println("Краулер: Успешное подключение к базе данных.")

	fetcher := crawler.NewHTTPFetcher(10 * time.Second)
	app := crawler.NewCrawler(5, 10, db, fetcher)

	go app.Start(ctx, []string{"https://golang.org"})
	log.Println("Краулер запущен. Нажмите CTRL+C для остановки.")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения, остановка краулера...")
	app.Stop()
	log.Println("Краулер успешно остановлен.")
}
