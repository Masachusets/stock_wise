package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/Masachusets/stock_wise/gen/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/log"
)

type WaybillHandlers struct {
	tpl  *template.Template
	pool *pgxpool.Pool
	svc  waybills.Service
}

func NewWaybillHandlers(tpl *template.Template, pool *pgxpool.Pool, svc waybills.Service) *WaybillHandlers {
	return &WaybillHandlers{tpl: tpl, pool: pool, svc: svc}
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

	rows, err := h.pool.Query(r.Context(), `SELECT w.id, w.number, w.issue_date::text, w.status,
		fd.name, td.name
	FROM waybills w
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE w.deleted_at IS NULL
	ORDER BY w.issue_date DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type waybillRow struct {
		ID        int32
		Number    string
		IssueDate string
		Status    string
		FromName  *string
		ToName    *string
	}
	var waybillsList []waybillRow
	for rows.Next() {
		var wb waybillRow
		if err := rows.Scan(&wb.ID, &wb.Number, &wb.IssueDate, &wb.Status, &wb.FromName, &wb.ToName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		waybillsList = append(waybillsList, wb)
	}

	data := map[string]interface{}{
		"Title":    "Накладные",
		"Active":   "waybills",
		"Waybills": waybillsList,
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

	var wb struct {
		Number    string
		IssueDate string
		Status    string
		FromName  *string
		ToName    *string
	}
	err = h.pool.QueryRow(r.Context(), `SELECT w.number, w.issue_date::text, w.status,
		fd.name, td.name
	FROM waybills w
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE w.id = $1`, id).Scan(&wb.Number, &wb.IssueDate, &wb.Status, &wb.FromName, &wb.ToName)
	if err != nil {
		http.Error(w, "накладная не найдена", http.StatusNotFound)
		return
	}

	eqRows, err := h.pool.Query(r.Context(), `SELECT DISTINCT e.inventory_number, COALESCE(e.model_name, ''), n.name
	FROM equipments_assignments a
	JOIN equipments e ON a.equipment_id = e.id
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE a.waybill_id = $1`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer eqRows.Close()

	type eqItem struct {
		InventoryNumber string
		ModelName       string
		Nomenclature    string
	}
	var equipmentsList []eqItem
	for eqRows.Next() {
		var item eqItem
		var modelName, nomenclature interface{}
		if err := eqRows.Scan(&item.InventoryNumber, &modelName, &nomenclature); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if modelName != nil {
			item.ModelName = fmt.Sprintf("%v", modelName)
		}
		if nomenclature != nil {
			item.Nomenclature = fmt.Sprintf("%v", nomenclature)
		}
		equipmentsList = append(equipmentsList, item)
	}

	data := map[string]interface{}{
		"Title":      fmt.Sprintf("Накладная %s", wb.Number),
		"Active":     "waybills",
		"Waybill":    wb,
		"WaybillID":  id,
		"Equipments": equipmentsList,
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
