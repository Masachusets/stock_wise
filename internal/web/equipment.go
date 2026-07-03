package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/Masachusets/stock_wise/gen/equipments"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/log"
)

type EquipmentHandlers struct {
	tpl  *template.Template
	pool *pgxpool.Pool
	svc  equipments.Service
}

func NewEquipmentHandlers(tpl *template.Template, pool *pgxpool.Pool, svc equipments.Service) *EquipmentHandlers {
	return &EquipmentHandlers{tpl: tpl, pool: pool, svc: svc}
}

func (h *EquipmentHandlers) Register(mux *http.ServeMux) {
	mux.HandleFunc("/equipments", h.list)
	mux.HandleFunc("/equipments/", h.detail)
	mux.HandleFunc("/equipments/add", h.add)
	mux.HandleFunc("/equipments/update", h.update)
	mux.HandleFunc("/equipments/deleted", h.deleted)
}

func (h *EquipmentHandlers) list(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/equipments" {
		http.NotFound(w, r)
		return
	}

	res, err := h.svc.List(r.Context(), &equipments.ListPayload{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := h.pool.Query(r.Context(), "SELECT id, code, name FROM nomenclatures ORDER BY code")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type nomItem struct {
		ID   int32
		Code string
		Name string
	}
	var nomenclatures []nomItem
	for rows.Next() {
		var n nomItem
		if err := rows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nomenclatures = append(nomenclatures, n)
	}

	data := map[string]interface{}{
		"Title":        "Оборудование",
		"Active":       "equipments",
		"Equipments":   res.Equipments,
		"Nomenclatures": nomenclatures,
	}
	RenderPage(w, h.tpl, "equipmentList", data)
}

func (h *EquipmentHandlers) detail(w http.ResponseWriter, r *http.Request) {
	invNum := strings.TrimPrefix(r.URL.Path, "/equipments/")
	if invNum == "" {
		http.NotFound(w, r)
		return
	}

	if r.Method == "DELETE" {
		if err := h.svc.Delete(r.Context(), &equipments.DeletePayload{InventoryNumber: invNum}); err != nil {
			log.Errorf(r.Context(), err, "delete equipment %s", invNum)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	res, err := h.svc.Get(r.Context(), &equipments.GetPayload{InventoryNumber: invNum})
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	rows, err := h.pool.Query(r.Context(), "SELECT id, code, name FROM nomenclatures ORDER BY code")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type nomItem struct {
		ID   int32
		Code string
		Name string
	}
	var nomenclatures []nomItem
	for rows.Next() {
		var n nomItem
		if err := rows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		nomenclatures = append(nomenclatures, n)
	}

	data := map[string]interface{}{
		"Title":        res.InventoryNumber,
		"Active":       "equipments",
		"Equipment":    res,
		"Nomenclatures": nomenclatures,
	}
	RenderPage(w, h.tpl, "equipmentDetail", data)
}

func (h *EquipmentHandlers) add(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		InventoryNumber string  `json:"inventory_number"`
		ModelName       string  `json:"model_name"`
		Status          string  `json:"status"`
		NomenclatureID  *int32  `json:"nomenclature_id"`
		SerialNumber    *string `json:"serial_number"`
		ManufactureDate *string `json:"manufacture_date"`
		ArrivalDate     *string `json:"arrival_date"`
		Notes           *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	payload := &equipments.CreateEquipmentPayload{
		InventoryNumber: body.InventoryNumber,
		ModelName:       body.ModelName,
		Status:          body.Status,
		NomenclatureID:  body.NomenclatureID,
		SerialNumber:    body.SerialNumber,
		ManufactureDate: normalizeDate(body.ManufactureDate),
		ArrivalDate:     normalizeDate(body.ArrivalDate),
		Notes:           body.Notes,
	}

	_, err := h.svc.Create(r.Context(), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var eqID int32
	err = h.pool.QueryRow(r.Context(), "SELECT id FROM equipments WHERE inventory_number = $1", body.InventoryNumber).Scan(&eqID)
	if err == nil {
		deptCode := 100
		h.pool.Exec(r.Context(),
			`INSERT INTO equipments_assignments (equipment_id, target_type, department_code)
			 VALUES ($1, 'warehouse', $2)`, eqID, deptCode)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *EquipmentHandlers) update(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		InventoryNumber string  `json:"inventory_number"`
		ModelName       *string `json:"model_name"`
		Status          *string `json:"status"`
		NomenclatureID  *int32  `json:"nomenclature_id"`
		SerialNumber    *string `json:"serial_number"`
		ManufactureDate *string `json:"manufacture_date"`
		ArrivalDate     *string `json:"arrival_date"`
		Notes           *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	payload := &equipments.UpdateEquipmentPayload{
		InventoryNumber: body.InventoryNumber,
		ModelName:       body.ModelName,
		Status:          body.Status,
		NomenclatureID:  body.NomenclatureID,
		SerialNumber:    body.SerialNumber,
		ManufactureDate: normalizeDate(body.ManufactureDate),
		ArrivalDate:     normalizeDate(body.ArrivalDate),
		Notes:           body.Notes,
	}
	if _, err := h.svc.Update(r.Context(), payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *EquipmentHandlers) deleted(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/equipments/deleted" {
		http.NotFound(w, r)
		return
	}
	rows, err := h.pool.Query(r.Context(), `SELECT
		e.inventory_number, COALESCE(e.model_name, ''),
		n.code, n.name, e.status, e.deleted_at::text
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE e.deleted_at IS NOT NULL
	ORDER BY e.deleted_at DESC`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type deletedItem struct {
		InventoryNumber string
		ModelName       string
		NomCode         string
		NomName         string
		Status          string
		DeletedAt       string
	}
	var items []deletedItem
	for rows.Next() {
		var item deletedItem
		if err := rows.Scan(&item.InventoryNumber, &item.ModelName, &item.NomCode, &item.NomName, &item.Status, &item.DeletedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	data := map[string]interface{}{
		"Title": "Удалённое оборудование",
		"Active": "equipments",
		"Items":  items,
	}
	RenderPage(w, h.tpl, "equipmentDeleted", data)
}
