package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Masachusets/stock_wise/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/log"
)
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// runApp инициализирует и запускает HTTP-сервер StockWise.
func runApp(cfg *config.Config) error {
	// Настройка логирования
	var format log.FormatFunc
	switch cfg.Log {
	case "json":
		format = log.FormatJSON
	case "terminal":
		format = log.FormatTerminal
	default:
		format = log.FormatText
	}

	ctx := log.Context(context.Background(), log.WithFormat(format))
	ctx, cancel := context.WithCancel(ctx)

	if cfg.Debug {
		ctx = log.Context(context.Background(), log.WithFormat(format), log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}
	
	log.Printf(ctx, "starting server on port: %v", cfg.Port)

	// Создаем канал, используемый обработчиком сигналов и 
	// серверными горутинами для уведомления главной горутины 
	// о необходимости остановки сервера.
	errCh := make(chan error, 1)

	// Подключение к PostgreSQL
	pool, err := pgxpool.New(ctx, cfg.DB)
	if err != nil {
		log.Fatalf(ctx, err, "db connect: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Errorf(ctx, err, "db ping: %v", err)
		errCh <- err
	}
	log.Printf(ctx, "connected to PostgreSQL")

	var wg sync.WaitGroup

	handleHTTPServer(ctx, cfg, &wg, pool, errCh)

	// Ожидание сигнала или ошибки сервера
	select {
	case <-ctx.Done():
		log.Printf(ctx, "shutting down server")
	case err := <-errCh:
		fmt.Printf("server failed: %v", err)
	}

	// Отправка сигнал отмены горутинам.
	cancel()
	
	wg.Wait()
	log.Printf(ctx, "server stopped")
	return nil
}
