package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cis-engine/internal/indexer"
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
	log.Println("Индексатор: Успешное подключение к базе данных.")

	app := indexer.NewIndexer(db, 2*time.Second)

	go app.Start(ctx)

	log.Println("Индексатор запущен. Нажмите CTRL+C для остановки.")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения, остановка индексатора...")
	app.Stop()
	log.Println("Индексатор успешно остановлен.")
}
