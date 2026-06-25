package importer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xuri/excelize/v2"
)

type equipmentRow struct {
	inventoryNumber string
	serialNumber    string
	nomenclatureID  int
	modelName       string
	manufactureDate string
	arrivalDate     string
	status          string
	formNumber      string
	location        string
	notes           string
}

func (imp *Importer) importEquipments(ctx context.Context) error {
	path := imp.excelPath("UMS.xlsx")
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("Учет")
	if err != nil {
		return err
	}

	// Build nomenclature lookup map
	nomMap, err := imp.buildNomenclatureMap(ctx)
	if err != nil {
		return fmt.Errorf("building nomenclature map: %w", err)
	}

	var items []equipmentRow
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 2 {
			continue
		}

		invNum := cellString(getCol(row, 1)) // B
		if invNum == "" {
			continue
		}

		item := equipmentRow{
			inventoryNumber: invNum,
			serialNumber:    cellString(getCol(row, 3)), // D
			modelName:       cellString(getCol(row, 15)), // P = Модель
			status:          statusMap(cellString(getCol(row, 10))), // K
			formNumber:      cellString(getCol(row, 11)), // L
			location:        cellString(getCol(row, 14)), // O
			notes:           cellString(getCol(row, 19)), // T
		}

		// Nomenclature ID
		nomCode := cellString(getCol(row, 5)) // F
		if id, ok := nomMap[nomCode]; ok {
			item.nomenclatureID = id
		}

		// Manufacture date (MM.YYYY or YYYY → DATE)
		dateStr := cellString(getCol(row, 6)) // G = Год выпуска
		if dateStr != "" {
			item.manufactureDate = parseManufactureDate(dateStr)
		}

		// Arrival date (MM.YYYY → DATE)
		dateStr = cellString(getCol(row, 7)) // H
		if t, err := time.Parse("01.2006", dateStr); err == nil {
			item.arrivalDate = t.Format("2006-01-02")
		}

		items = append(items, item)
	}

	log.Printf("equipments: найдено %d записей", len(items))

	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx, `
			INSERT INTO equipments (inventory_number, serial_number, nomenclature_id,
				model_name, manufacture_date, arrival_date, status, form_number, location, notes)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			ON CONFLICT (inventory_number) DO NOTHING`,
			item.inventoryNumber,
			nullStr(item.serialNumber),
			nullInt(item.nomenclatureID),
			nullStr(item.modelName),
			nullStr(item.manufactureDate),
			nullStr(item.arrivalDate),
			item.status,
			nullStr(item.formNumber),
			nullStr(item.location),
			nullStr(item.notes),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func parseManufactureDate(raw string) string {
	s := raw
	// Strip parenthetical suffixes like (03.2015)
	if idx := indexOf(s, "("); idx >= 0 {
		s = s[:idx]
	}
	s = trimSpace(s)

	// Try MM.YYYY
	if t, err := time.Parse("01.2006", s); err == nil {
		return t.Format("2006-01-02")
	}
	// Try YYYY
	if len(s) == 4 {
		if t, err := time.Parse("2006", s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return ""
}

func indexOf(s, c string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c[0] {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func (imp *Importer) buildNomenclatureMap(ctx context.Context) (map[string]int, error) {
	rows, err := imp.db.Query(ctx, `SELECT id, code FROM nomenclatures`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var id int
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		m[code] = id
	}
	return m, nil
}

func getCol(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullInt(i int) interface{} {
	if i == 0 {
		return nil
	}
	return i
}
