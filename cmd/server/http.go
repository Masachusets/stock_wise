package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Masachusets/stock_wise/internal/api"
	"github.com/Masachusets/stock_wise/internal/config"
	"github.com/Masachusets/stock_wise/internal/web"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

func handleHTTPServer(
	ctx context.Context,
	cfg *config.Config,
	wg *sync.WaitGroup,
	pool *pgxpool.Pool,
	errc chan error,
) {
	// Создание сервисов (internal/*)
	services := NewServices(pool)

	// Создание HTTP-сервера
	rootMux := http.NewServeMux()

	// Регистрация API-хэндлеров (goa)
	apiMux := goahttp.NewMuxer()
	if cfg.Debug {
		debug.MountPprofHandlers(debug.Adapt(apiMux))
		debug.MountDebugLogEnabler(debug.Adapt(apiMux))
	}
	api.RegisterRoutes(
		apiMux,
		services.Nomenclatures,
		services.Departments,
		services.Cards,
		services.Equipments,
		services.Waybills,
		services.Assignments,
	)

	// Регистрация веб-хэндлеров (шаблоны + статика)
	tpl := web.LoadTemplates()
	web.RegisterRoutes(rootMux, tpl, pool, services.Equipments, services.WaybillsSvc)

	// Монтирование API на /api/
	rootMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	var handler http.Handler = rootMux
	if cfg.Debug {
		handler = debug.HTTP()(handler)
	}
	handler = log.HTTP(ctx)(handler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 60 * time.Second,
	}

	(*wg).Add(1)
	go func() {
		defer (*wg).Done()

		go func() {
			log.Printf(ctx, "HTTP server listening on %d", cfg.Port)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		log.Printf(ctx, "shutting down HTTP server at %d", cfg.Port)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf(ctx, "shutdown error: %v", err)
		}
	}()
}
