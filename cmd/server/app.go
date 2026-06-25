package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	assignments "github.com/Masachusets/stock_wise/gen/assignments"
	cards "github.com/Masachusets/stock_wise/gen/cards"
	departments "github.com/Masachusets/stock_wise/gen/departments"
	equipments "github.com/Masachusets/stock_wise/gen/equipments"
	nomenclatures "github.com/Masachusets/stock_wise/gen/nomenclatures"
	waybills "github.com/Masachusets/stock_wise/gen/waybills"
	assignmentssvr "github.com/Masachusets/stock_wise/gen/http/assignments/server"
	cardssvr "github.com/Masachusets/stock_wise/gen/http/cards/server"
	departmentssvr "github.com/Masachusets/stock_wise/gen/http/departments/server"
	equipmentssvr "github.com/Masachusets/stock_wise/gen/http/equipments/server"
	nomenclaturessvr "github.com/Masachusets/stock_wise/gen/http/nomenclatures/server"
	waybillssvr "github.com/Masachusets/stock_wise/gen/http/waybills/server"
	"github.com/Masachusets/stock_wise/internal/config"
	svcassignments "github.com/Masachusets/stock_wise/internal/assignments"
	svccards "github.com/Masachusets/stock_wise/internal/cards"
	svcdepartments "github.com/Masachusets/stock_wise/internal/departments"
	svcequipments "github.com/Masachusets/stock_wise/internal/equipments"
	svcnomenclatures "github.com/Masachusets/stock_wise/internal/nomenclatures"
	svcwaybills "github.com/Masachusets/stock_wise/internal/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
	goahttp "goa.design/goa/v3/http"
)

// loggingHandler — HTTP-мидлвейнер для логирования запросов.
type loggingHandler struct {
	handler http.Handler
}

func (h *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
	h.handler.ServeHTTP(sw, r)
	duration := time.Since(start).String()

	if sw.status >= 500 {
		slog.Error("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", duration,
		)
	} else if sw.status >= 400 {
		slog.Warn("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", duration,
		)
	} else {
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration", duration,
		)
	}
}

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
	// Настройка уровня логирования
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Log)); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	slog.Info("starting server", "port", cfg.Port)

	// Подключение к PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DB)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	slog.Info("connected to PostgreSQL")

	// Создание сервисов (internal/*)
	nomenclaturesSvc := svcnomenclatures.New(pool)
	departmentsSvc := svcdepartments.New(pool)
	cardsSvc := svccards.New(pool)
	equipmentsSvc := svcequipments.New(pool)
	waybillsSvc := svcwaybills.New(pool)
	assignmentsSvc := svcassignments.New(pool)

	// Создание endpoints (gen/*)
	nomenclaturesEndpoints := nomenclatures.NewEndpoints(nomenclaturesSvc)
	departmentsEndpoints := departments.NewEndpoints(departmentsSvc)
	cardsEndpoints := cards.NewEndpoints(cardsSvc)
	equipmentsEndpoints := equipments.NewEndpoints(equipmentsSvc)
	waybillsEndpoints := waybills.NewEndpoints(waybillsSvc)
	assignmentsEndpoints := assignments.NewEndpoints(assignmentsSvc)

	// Настройка HTTP-транспорта
	mux := goahttp.NewMuxer()
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	// Обработчик ошибок Goa — логирует ошибки с контекстом запроса
	eh := func(ctx context.Context, w http.ResponseWriter, err error) {
		slog.Error("endpoint error", "error", err)
	}

	// Создание HTTP-хэндлеров
	nomSvr := nomenclaturessvr.New(nomenclaturesEndpoints, mux, dec, enc, eh, nil)
	deptSvr := departmentssvr.New(departmentsEndpoints, mux, dec, enc, eh, nil)
	cardsSvr := cardssvr.New(cardsEndpoints, mux, dec, enc, eh, nil)
	eqSvr := equipmentssvr.New(equipmentsEndpoints, mux, dec, enc, eh, nil)
	wbSvr := waybillssvr.New(waybillsEndpoints, mux, dec, enc, eh, nil)
	asnSvr := assignmentssvr.New(assignmentsEndpoints, mux, dec, enc, eh, nil)

	// Монтирование маршрутов
	nomenclaturessvr.Mount(mux, nomSvr)
	departmentssvr.Mount(mux, deptSvr)
	cardssvr.Mount(mux, cardsSvr)
	equipmentssvr.Mount(mux, eqSvr)
	waybillssvr.Mount(mux, wbSvr)
	assignmentssvr.Mount(mux, asnSvr)

	// Вывод смонтированных маршрутов
	for _, m := range nomSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}
	for _, m := range deptSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}
	for _, m := range cardsSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}
	for _, m := range eqSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}
	for _, m := range wbSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}
	for _, m := range asnSvr.Mounts {
		slog.Debug("route mounted", "method", m.Method, "verb", m.Verb, "pattern", m.Pattern)
	}

	// Создание HTTP-сервера
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           &loggingHandler{handler: mux},
		ReadHeaderTimeout: 60 * time.Second,
	}

	// Обработка сигналов (SIGINT/SIGTERM) для graceful shutdown
	errc := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// Запуск HTTP-сервера
	go func() {
		slog.Info("server listening", "port", cfg.Port)
		errc <- srv.ListenAndServe()
	}()

	// Ожидание сигнала или ошибки сервера
	<-errc
	slog.Info("shutting down server")

	// Graceful shutdown с таймаутом 30 секунд
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	slog.Info("server stopped")
	return nil
}
