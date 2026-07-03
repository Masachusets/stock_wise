package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Masachusets/stock_wise/gen/waybills"
	svcwaybills "github.com/Masachusets/stock_wise/internal/waybills"
	"goa.design/clue/log"
)

type WaybillHandlers struct {
	tpl *template.Template
	svc svcwaybills.WebService
}

func NewWaybillHandlers(tpl *template.Template, svc svcwaybills.WebService) *WaybillHandlers {
	return &WaybillHandlers{tpl: tpl, svc: svc}
}

func (h *WaybillHandlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/waybills", h.list)
	mux.HandleFunc("/waybills/", h.detail)
	mux.HandleFunc("/api/waybills/", h.api)
}

func (h *WaybillHandlers) list(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/waybills" {
		http.NotFound(w, r)
		return
	}

	items, err := h.svc.ListForWeb(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":    "Накладные",
		"Active":   "waybills",
		"Waybills": items,
	}
	RenderPage(w, h.tpl, "waybillsPage", data)
}

func (h *WaybillHandlers) detail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/waybills/")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	detail, err := h.svc.GetForWeb(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "накладная не найдена", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{
		"Title":      fmt.Sprintf("Накладная %s", detail.Number),
		"Active":     "waybills",
		"Waybill":    detail,
		"WaybillID":  id,
		"Equipments": detail.Equipments,
	}
	RenderPage(w, h.tpl, "waybillDetail", data)
}

func (h *WaybillHandlers) api(w http.ResponseWriter, r *http.Request) {
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
		_, err := h.svc.Sign(r.Context(), &waybills.SignPayload{ID: int32(id)})
		if err != nil {
			log.Errorf(r.Context(), err, "sign waybill %d", id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(parts) == 2 && parts[1] == "archive" && r.Method == "POST" {
		_, err := h.svc.Archive(r.Context(), &waybills.ArchivePayload{ID: int32(id)})
		if err != nil {
			log.Errorf(r.Context(), err, "archive waybill %d", id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(parts) == 1 && r.Method == "DELETE" {
		err := h.svc.Delete(r.Context(), &waybills.DeletePayload{ID: int32(id)})
		if err != nil {
			log.Errorf(r.Context(), err, "delete waybill %d", id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	http.NotFound(w, r)
}
