package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Masachusets/stock_wise/gen/equipments"
	"github.com/Masachusets/stock_wise/gen/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
	"goa.design/clue/log"
)

// normalizeDate конвертирует "YYYY-MM" в "YYYY-MM-01" для PostgreSQL DATE.
func normalizeDate(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	v := *s
	if len(v) == 7 { // "YYYY-MM"
		v += "-01"
	}
	return &v
}

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

func registerWebHandlers(mux *http.ServeMux, tpl *template.Template, pool *pgxpool.Pool, eqSvc equipments.Service, wbSvc waybills.Service) {
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

		// Получить список номенклатур для формы
		nomRows, err := pool.Query(
			r.Context(), 
			"SELECT id, code, name FROM nomenclatures ORDER BY code",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer nomRows.Close()

		
		type nomItem struct {
			ID   int32
			Code string
			Name string
		}
		var nomenclatures []nomItem
		for nomRows.Next() {
			var n nomItem
			if err := nomRows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			nomenclatures = append(nomenclatures, n)
		}

		slog.Debug("equipment list", "count", len(res.Equipments))
		data := map[string]interface{}{
			"Title":        "Оборудование",
			"Active":       "equipments",
			"Equipments":   res.Equipments,
			"Nomenclatures": nomenclatures,
		}
		renderPage(w, tpl, "equipmentList", data)
	})

	mux.HandleFunc("/equipments/", func(w http.ResponseWriter, r *http.Request) {
		invNum := strings.TrimPrefix(r.URL.Path, "/equipments/")
		if invNum == "" {
			http.NotFound(w, r)
			return
		}

		// DELETE — удаление оборудования
		if r.Method == "DELETE" {
			if err := eqSvc.Delete(r.Context(), &equipments.DeletePayload{InventoryNumber: invNum}); err != nil {
				log.Errorf(r.Context(), err, "delete equipment %s", invNum)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// GET — просмотр карточки оборудования
		res, err := eqSvc.Get(r.Context(), &equipments.GetPayload{InventoryNumber: invNum})
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Получить список номенклатур для формы редактирования
		nomRows, err := pool.Query(r.Context(), "SELECT id, code, name FROM nomenclatures ORDER BY code")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer nomRows.Close()

		type nomItem struct {
			ID   int32
			Code string
			Name string
		}
		var nomenclatures []nomItem
		for nomRows.Next() {
			var n nomItem
			if err := nomRows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
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
		renderPage(w, tpl, "equipmentDetail", data)
	})

	mux.HandleFunc("/equipments/add", func(w http.ResponseWriter, r *http.Request) {
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

		slog.Debug("adding equipment", "inventory_number", body.InventoryNumber, "model_name", body.ModelName)

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

		_, err := eqSvc.Create(r.Context(), payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Автоматическое закрепление за складом
		var eqID int32
		err = pool.QueryRow(r.Context(), "SELECT id FROM equipments WHERE inventory_number = $1", body.InventoryNumber).Scan(&eqID)
		if err == nil {
			// Закрепить за складом (код 100)
			deptCode := 100
			pool.Exec(r.Context(),
				`INSERT INTO equipments_assignments (equipment_id, target_type, department_code)
				 VALUES ($1, 'warehouse', $2)`, eqID, deptCode)
		}

		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/equipments/deleted", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/equipments/deleted" {
			http.NotFound(w, r)
			return
		}
		rows, err := pool.Query(r.Context(), `SELECT
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
			"Title":   "Удалённое оборудование",
			"Active":  "equipments",
			"Items":   items,
		}
		renderPage(w, tpl, "equipmentDeleted", data)
	})

	mux.HandleFunc("/equipments/update", func(w http.ResponseWriter, r *http.Request) {
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
		if _, err := eqSvc.Update(r.Context(), payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/waybills", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/waybills" {
			http.NotFound(w, r)
			return
		}
		rows, err := pool.Query(r.Context(), `SELECT w.id, w.number, w.issue_date::text, w.status,
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
			"Title":   "Накладные",
			"Active":  "waybills",
			"Waybills": waybillsList,
		}
		renderPage(w, tpl, "waybillsPage", data)
	})

	mux.HandleFunc("/waybills/", func(w http.ResponseWriter, r *http.Request) {
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

		// Получить информацию о накладной
		var wb struct {
			Number    string
			IssueDate string
			Status    string
			FromName  *string
			ToName    *string
		}
		err = pool.QueryRow(r.Context(), `SELECT w.number, w.issue_date::text, w.status,
			fd.name, td.name
		FROM waybills w
		LEFT JOIN departments fd ON w.from_dept = fd.code
		LEFT JOIN departments td ON w.to_dept = td.code
		WHERE w.id = $1`, id).Scan(&wb.Number, &wb.IssueDate, &wb.Status, &wb.FromName, &wb.ToName)
		if err != nil {
			http.Error(w, "накладная не найдена", http.StatusNotFound)
			return
		}

		// Получить оборудование по накладной (через equipments_assignments)
		eqRows, err := pool.Query(r.Context(), `SELECT DISTINCT e.inventory_number, COALESCE(e.model_name, ''), n.name
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
		var equipments []eqItem
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
			equipments = append(equipments, item)
		}

		data := map[string]interface{}{
			"Title":      fmt.Sprintf("Накладная %s", wb.Number),
			"Active":     "waybills",
			"Waybill":    wb,
			"WaybillID":  id,
			"Equipments": equipments,
		}
		renderPage(w, tpl, "waybillDetail", data)
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
