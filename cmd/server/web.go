package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Masachusets/stock_wise/gen/equipments"
	"github.com/Masachusets/stock_wise/gen/waybills"
)

var funcMap = template.FuncMap{
	"statusLabel": func(s string) string {
		labels := map[string]string{
			"exp":         "Эксплуатируемое",
			"exp_int":     "Эксплуатируемое (интернет)",
			"exp_sp":      "Эксплуатируемое (категорир.)",
			"broken":      "Неисправное",
			"written_off": "Списанное",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
	"statusLabelWaybill": func(s string) string {
		labels := map[string]string{
			"draft":    "Черновик",
			"signed":   "Подписан",
			"archived": "В архиве",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
	"targetLabel": func(s string) string {
		labels := map[string]string{
			"employee":   "Сотрудник",
			"department": "Подразделение",
			"warehouse":  "Склад",
		}
		if l, ok := labels[s]; ok {
			return l
		}
		return s
	},
}

func loadTemplates() *template.Template {
	return template.Must(
		template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"),
	)
}

func registerWebHandlers(mux *http.ServeMux, tpl *template.Template, eqSvc equipments.Service, wbSvc waybills.Service) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/equipments", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/equipments", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/equipments" {
			http.NotFound(w, r)
			return
		}
		res, err := eqSvc.List(r.Context(), &equipments.ListPayload{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		slog.Debug("equipment list", "count", len(res.Equipments))
		data := map[string]interface{}{
			"Title":     "Оборудование",
			"Active":    "equipments",
			"Equipments": res.Equipments,
		}
		renderPage(w, tpl, "equipmentList", data)
	})

	mux.HandleFunc("/equipments/", func(w http.ResponseWriter, r *http.Request) {
		invNum := strings.TrimPrefix(r.URL.Path, "/equipments/")
		if invNum == "" {
			http.NotFound(w, r)
			return
		}
		res, err := eqSvc.Get(r.Context(), &equipments.GetPayload{InventoryNumber: invNum})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		data := map[string]interface{}{
			"Title":     res.InventoryNumber,
			"Active":    "equipments",
			"Equipment": res,
		}
		renderPage(w, tpl, "equipmentDetail", data)
	})

	mux.HandleFunc("/waybills", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/waybills" {
			http.NotFound(w, r)
			return
		}
		res, err := wbSvc.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Title":   "Накладные",
			"Active":  "waybills",
			"Waybills": res.Waybills,
		}
		renderPage(w, tpl, "waybillsPage", data)
	})

	// API endpoints for waybill actions (used by JS)
	mux.HandleFunc("/api/waybills/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/waybills/"), "/")
		if len(parts) < 1 {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		if len(parts) == 2 && parts[1] == "sign" && r.Method == "POST" {
			_, err := wbSvc.Sign(r.Context(), &waybills.SignPayload{ID: int32(id)})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if len(parts) == 2 && parts[1] == "archive" && r.Method == "POST" {
			_, err := wbSvc.Archive(r.Context(), &waybills.ArchivePayload{ID: int32(id)})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if len(parts) == 1 && r.Method == "DELETE" {
			err := wbSvc.Delete(r.Context(), &waybills.DeletePayload{ID: int32(id)})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		http.NotFound(w, r)
	})

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	slog.Info("web handlers registered")
}

func renderPage(w http.ResponseWriter, tpl *template.Template, pageTmpl string, data map[string]interface{}) {
	var buf strings.Builder
	if err := tpl.ExecuteTemplate(&buf, pageTmpl, data); err != nil {
		slog.Error("template error", "error", err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	data["Content"] = template.HTML(buf.String())
	if err := tpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		slog.Error("layout error", "error", err)
	}
}
