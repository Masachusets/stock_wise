package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	assignments "github.com/Masachusets/stock_wise/gen/assignments"
	cards "github.com/Masachusets/stock_wise/gen/cards"
	departments "github.com/Masachusets/stock_wise/gen/departments"
	equipments "github.com/Masachusets/stock_wise/gen/equipments"
	assignmentssvr "github.com/Masachusets/stock_wise/gen/http/assignments/server"
	cardssvr "github.com/Masachusets/stock_wise/gen/http/cards/server"
	departmentssvr "github.com/Masachusets/stock_wise/gen/http/departments/server"
	equipmentssvr "github.com/Masachusets/stock_wise/gen/http/equipments/server"
	nomenclaturessvr "github.com/Masachusets/stock_wise/gen/http/nomenclatures/server"
	waybillssvr "github.com/Masachusets/stock_wise/gen/http/waybills/server"
	nomenclatures "github.com/Masachusets/stock_wise/gen/nomenclatures"
	waybills "github.com/Masachusets/stock_wise/gen/waybills"
	svcassignments "github.com/Masachusets/stock_wise/internal/assignments"
	svccards "github.com/Masachusets/stock_wise/internal/cards"
	"github.com/Masachusets/stock_wise/internal/config"
	svcdepartments "github.com/Masachusets/stock_wise/internal/departments"
	svcequipments "github.com/Masachusets/stock_wise/internal/equipments"
	svcnomenclatures "github.com/Masachusets/stock_wise/internal/nomenclatures"
	svcwaybills "github.com/Masachusets/stock_wise/internal/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/debug"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

// Services содержит все бизнес-сервисы приложения.
type Services struct {
    Nomenclatures nomenclatures.Service
    Departments   departments.Service
    Cards         cards.Service
    Equipments    equipments.Service
    Waybills      waybills.Service
    Assignments   assignments.Service
}

// NewServices создаёт все сервисы с использованием переданного пула соединений.
func NewServices(pool *pgxpool.Pool) *Services {
	return &Services{
        Nomenclatures: svcnomenclatures.New(svcnomenclatures.NewPostgresRepository(pool)),
        Departments:   svcdepartments.New(svcdepartments.NewPostgresRepository(pool)),
        Cards:         svccards.New(svccards.NewPostgresRepository(pool)),
        Equipments:    svcequipments.New(svcequipments.NewPostgresRepository(pool)),
        Waybills:      svcwaybills.New(svcwaybills.NewPostgresRepository(pool)),
        Assignments:   svcassignments.New(svcassignments.NewPostgresRepository(pool)),
    }
}

// Endpoints содержит все goa-эндпоинты для HTTP.
type Endpoints struct {
    Nomenclatures *nomenclatures.Endpoints
    Departments   *departments.Endpoints
    Cards         *cards.Endpoints
    Equipments    *equipments.Endpoints
    Waybills      *waybills.Endpoints
    Assignments   *assignments.Endpoints
}

// NewEndpoints создаёт эндпоинты на основе сервисов.
func NewEndpoints(svc *Services) *Endpoints {
    return &Endpoints{
        Nomenclatures: nomenclatures.NewEndpoints(svc.Nomenclatures),
        Departments:   departments.NewEndpoints(svc.Departments),
        Cards:         cards.NewEndpoints(svc.Cards),
        Equipments:    equipments.NewEndpoints(svc.Equipments),
        Waybills:      waybills.NewEndpoints(svc.Waybills),
        Assignments:   assignments.NewEndpoints(svc.Assignments),
    }
}

func handleHTTPServer(
	ctx context.Context,
	cfg *config.Config,
	wg *sync.WaitGroup,
	pool *pgxpool.Pool,
	errc chan error, 
) {
	// Предоставьте декодер запроса и кодировщик ответа, специфичные для данного транспорта.
	// Пакет goa http имеет встроенную поддержку JSON, XML и gob.
	// Другие кодировки можно использовать, предоставив соответствующие функции,
	// см. goa.design/implement/encoding.
	var (
		dec = goahttp.RequestDecoder
		enc = goahttp.ResponseEncoder
	)
	// Создайте мультиплексор HTTP-запросов для сервиса и подключите
	// конечные точки отладки и профилирования в режиме отладки.
	var mux goahttp.Muxer
	mux = goahttp.NewMuxer()
	if cfg.Debug {
		// Для профилирования памяти смонтируйте обработчики pprof в каталог /debug/pprof.
		debug.MountPprofHandlers(debug.Adapt(mux))
		// Подключите конечную точку /debug, чтобы включить или отключить 
		// ведение отладочных журналов во время выполнения.
		debug.MountDebugLogEnabler(debug.Adapt(mux))
	}
	
	// Создание сервисов (internal/*)
	services := NewServices(pool)

	// Создание endpoints (gen/*)
	endpoints := NewEndpoints(services)

	// Создание обработчика ошибок
	eh := errorHandler(ctx)

	// Оберните конечные точки в специфичные для транспорта уровни.
	// Сгенерированные серверные пакеты содержат код, созданный на основе проекта, 
	// который сопоставляет входные и выходные структуры данных сервиса с 
	// HTTP-запросами и ответами.	
	nomSvr    := nomenclaturessvr.New(endpoints.Nomenclatures, mux, dec, enc, eh, nil)
	deptSvr   := departmentssvr.New(endpoints.Departments, mux, dec, enc, eh, nil)
	cardsSvr  := cardssvr.New(endpoints.Cards, mux, dec, enc, eh, nil)
	equipSvr  := equipmentssvr.New(endpoints.Equipments, mux, dec, enc, eh, nil)
	wbSvr     := waybillssvr.New(endpoints.Waybills, mux, dec, enc, eh, nil)
	assignSvr := assignmentssvr.New(endpoints.Assignments, mux, dec, enc, eh, nil)

	// Монтирование маршрутов
	nomenclaturessvr.Mount(mux, nomSvr)
	departmentssvr.Mount(mux, deptSvr)
	cardssvr.Mount(mux, cardsSvr)
	equipmentssvr.Mount(mux, equipSvr)
	waybillssvr.Mount(mux, wbSvr)
	assignmentssvr.Mount(mux, assignSvr)

	// Создание HTTP-сервера
	rootMux := http.NewServeMux()

	// Регистрация веб-хэндлеров (шаблоны + статика)
	templates := loadTemplates()
	registerWebHandlers(
		rootMux, 
		templates, 
		pool,
		services.Equipments, 
		services.Waybills,
	)

	// Монтирование API-хэндлеров на /api/
	rootMux.Handle("/api/", http.StripPrefix("/api", mux))

	var handler http.Handler = rootMux
	if cfg.Debug {
		// Если включена отладки, записывает в журнал тела запросов и ответов.
		handler = debug.HTTP()(handler)
	}
	handler = log.HTTP(ctx)(handler)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 60 * time.Second,
	}

	(*wg).Add(1)
	go func () {
		defer (*wg).Done()

		// Запуск HTTP-сервера в отдельной горутине
		go func() {
			log.Printf(ctx, "HTTP server listening on %d", cfg.Port)
			errc <- srv.ListenAndServe()
		}()

		<-ctx.Done()
		log.Printf(ctx, "shutting down HTTP server at %d", cfg.Port)

		// Graceful shutdown с задержкой
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf(ctx, "shutdown error: %v", err)
		}
	}()
}

// errorHandler возвращает функцию, которая записывает и регистрирует заданную ошибку.
func errorHandler(logCtx context.Context) func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		log.Errorf(logCtx, err, "request error: %v", err)
	}
}