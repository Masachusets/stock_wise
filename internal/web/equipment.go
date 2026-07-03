package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	svcequipments "github.com/Masachusets/stock_wise/internal/equipments"
	"goa.design/clue/log"
)

type EquipmentHandlers struct {
	tpl *template.Template
	svc *svcequipments.Service
}

func NewEquipmentHandlers(tpl *template.Template, svc *svcequipments.Service) *EquipmentHandlers {
	return &EquipmentHandlers{tpl: tpl, svc: svc}
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

	items, err := h.svc.ListForWeb(r.Context(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	nomenclatures, err := h.svc.ListNomenclatures(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":        "Оборудование",
		"Active":       "equipments",
		"Equipments":   items,
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

	// DELETE — удаление оборудования
	if r.Method == "DELETE" {
		if err := h.svc.DeleteByInvNum(r.Context(), invNum); err != nil {
			log.Errorf(r.Context(), err, "delete equipment %s", invNum)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// GET — просмотр карточки оборудования
	detail, err := h.svc.GetForWeb(r.Context(), invNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	nomenclatures, err := h.svc.ListNomenclatures(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":        detail.InventoryNumber,
		"Active":       "equipments",
		"Equipment":    detail,
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

	eq := &svcequipments.Equipment{
		InventoryNumber: body.InventoryNumber,
		ModelName:       body.ModelName,
		Status:          body.Status,
		NomenclatureID:  body.NomenclatureID,
		SerialNumber:    body.SerialNumber,
		ManufactureDate: normalizeDate(body.ManufactureDate),
		ArrivalDate:     normalizeDate(body.ArrivalDate),
		Notes:           body.Notes,
	}

	if err := h.svc.CreateWithAssignment(r.Context(), eq, 100); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

	eq := &svcequipments.Equipment{
		InventoryNumber: body.InventoryNumber,
		ModelName:       derefStr(body.ModelName),
		Status:          derefStr(body.Status),
		NomenclatureID:  body.NomenclatureID,
		SerialNumber:    body.SerialNumber,
		ManufactureDate: normalizeDate(body.ManufactureDate),
		ArrivalDate:     normalizeDate(body.ArrivalDate),
		Notes:           body.Notes,
	}

	if err := h.svc.UpdateByDomain(r.Context(), eq); err != nil {
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

	items, err := h.svc.ListDeleted(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":  "Удалённое оборудование",
		"Active": "equipments",
		"Items":  items,
	}
	RenderPage(w, h.tpl, "equipmentDeleted", data)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
