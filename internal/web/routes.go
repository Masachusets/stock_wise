package web

import (
	"html/template"
	"net/http"

	"github.com/Masachusets/stock_wise/gen/equipments"
	"github.com/Masachusets/stock_wise/gen/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(mux *http.ServeMux, tpl *template.Template, pool *pgxpool.Pool, eqSvc equipments.Service, wbSvc waybills.Service) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/equipments", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	eqHandlers := NewEquipmentHandlers(tpl, pool, eqSvc)
	eqHandlers.Register(mux)

	wbHandlers := NewWaybillHandlers(tpl, pool, wbSvc)
	wbHandlers.Register(mux)

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
}
