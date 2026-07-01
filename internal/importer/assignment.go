package importer

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type assignmentRow struct {
	equipmentID     int
	inventoryNumber string
	targetType      string
	cardNumber      *int
	departmentCode  *int
	comment         string
}

type waybillInfo struct {
	number string
	toDept string
}

func (imp *Importer) importAssignments(ctx context.Context) error {
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

	// Build lookup maps
	eqMap, err := imp.buildEquipmentMap(ctx)
	if err != nil {
		return err
	}
	empMap, err := imp.buildEmployeeMap(ctx)
	if err != nil {
		return err
	}

	deptWaybillMap, err := imp.buildDeptWaybillMap(ctx)
	if err != nil {
		return err
	}

	// Still need waybills from Excel for other types
	deptNameToCode, err := imp.buildDepartmentNameToCodeMap(ctx)
	if err != nil {
		return err
	}

	waybillData := make(map[string]waybillInfo)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 14 {
			continue
		}
		colM := cellString(getCol(row, 12))
		colN := cellString(getCol(row, 13))
		if colN == "" {
			continue
		}
		if cardNum, err := parseCardNumber(colN); err == nil {
			if _, exists := empMap[cardNum]; exists {
				continue
			}
		}
		if colM != "" {
			_, deptName := parseDeptTypeAndName(colM)
			if deptCode, exists := deptNameToCode[deptName]; exists {
				waybillData[colN] = waybillInfo{number: colN, toDept: deptCode}
			}
		}
	}

	_, err = imp.createWaybills(ctx, waybillData)
	if err != nil {
		return err
	}

	// Create assignments
	var items []assignmentRow
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
		eqID, ok := eqMap[invNum]
		if !ok {
			continue
		}

		colM := cellString(getCol(row, 12)) // M = Закреплены
		colN := cellString(getCol(row, 13)) // N = Номер карточки/накладной
		colT := cellString(getCol(row, 19)) // T = Примечание

		item := assignmentRow{
			equipmentID:     eqID,
			inventoryNumber: invNum,
			comment:         colT,
		}

		if colM == "!" {
			colM = "СКЛАД"
		}

		if colM != "" {
			typ, deptName := parseDeptTypeAndName(colM)

			if typ == "pogk" || typ == "pogz" || typ == "opk" {
				item.targetType = "department"
				if deptCode, exists := deptNameToCode[deptName]; exists {
					code, _ := strconv.Atoi(deptCode)
					item.departmentCode = &code
				}
			} else if isPersonName(colM) {
				if colN != "" {
					if cardNum, err := parseCardNumber(colN); err == nil {
						if _, exists := empMap[cardNum]; exists {
							item.targetType = "employee"
							item.cardNumber = &cardNum
						}
					}
				}
			} else if typ == "warehouse" {
				item.targetType = "warehouse"
				for codeStr := range deptWaybillMap {
					code, _ := strconv.Atoi(codeStr)
					if code >= 100 && code <= 199 {
						item.departmentCode = &code
						break
					}
				}
			}
		}

		if item.targetType == "" {
			item.targetType = "warehouse"
			for codeStr := range deptWaybillMap {
				code, _ := strconv.Atoi(codeStr)
				if code >= 100 && code <= 199 {
					item.departmentCode = &code
					break
				}
			}
		}

		items = append(items, item)
	}

	log.Printf("assignments: найдено %d записей", len(items))

	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		var deptCode interface{}
		if item.departmentCode != nil {
			deptCode = *item.departmentCode
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO equipments_assignments (equipment_id, target_type, card_number, department_code, operator_comment, is_active)
			VALUES ($1, $2, $3, $4, $5, TRUE)`,
			item.equipmentID,
			item.targetType,
			item.cardNumber,
			deptCode,
			nullStr(item.comment),
		)
		if err != nil {
			log.Printf("assignment insert error for %s: targetType=%s card=%v dept=%v err=%v",
				item.inventoryNumber, item.targetType, item.cardNumber, deptCode, err)
			return err
		}
	}

	return tx.Commit(ctx)
}

func (imp *Importer) createWaybills(ctx context.Context, data map[string]waybillInfo) (map[string]int, error) {
	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	result := make(map[string]int)
	for num, info := range data {
		_, err := tx.Exec(ctx,
			`INSERT INTO waybills (number, issue_date, status, from_dept, to_dept)
			 VALUES ($1, CURRENT_DATE, 'archived', $2, $3)
			 ON CONFLICT (number) DO NOTHING`, num, "100", info.toDept,
		)
		if err != nil {
			log.Printf("waybill insert error for %q: %v", num, err)
			continue
		}

		var id int
		err = tx.QueryRow(ctx, `SELECT id FROM waybills WHERE number = $1`, num).Scan(&id)
		if err != nil {
			log.Printf("waybill lookup error for %q: %v", num, err)
			continue
		}
		result[num] = id
	}

	return result, tx.Commit(ctx)
}

func (imp *Importer) buildEquipmentMap(ctx context.Context) (map[string]int, error) {
	rows, err := imp.db.Query(ctx, `SELECT id, inventory_number FROM equipments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var id int
		var inv string
		if err := rows.Scan(&id, &inv); err != nil {
			return nil, err
		}
		m[inv] = id
	}
	return m, nil
}

func (imp *Importer) buildEmployeeMap(ctx context.Context) (map[int]bool, error) {
	rows, err := imp.db.Query(ctx, `SELECT number FROM cards`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int]bool)
	for rows.Next() {
		var cn int
		if err := rows.Scan(&cn); err != nil {
			return nil, err
		}
		m[cn] = true
	}
	return m, nil
}

func (imp *Importer) buildDepartmentMap(ctx context.Context) (map[string]int, error) {
	rows, err := imp.db.Query(ctx, `SELECT id, name FROM departments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[name] = id
	}
	return m, nil
}

func (imp *Importer) buildDepartmentNameToCodeMap(ctx context.Context) (map[string]string, error) {
	rows, err := imp.db.Query(ctx, `SELECT code, name FROM departments`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var code int
		var name string
		if err := rows.Scan(&code, &name); err != nil {
			return nil, err
		}
		m[name] = strconv.Itoa(code)
	}
	return m, nil
}

func (imp *Importer) buildDeptWaybillMap(ctx context.Context) (map[string]int, error) {
	rows, err := imp.db.Query(ctx, `SELECT id, number FROM waybills WHERE number LIKE 'ПОДР-%'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]int)
	for rows.Next() {
		var id int
		var number string
		if err := rows.Scan(&id, &number); err != nil {
			return nil, err
		}
		code := strings.TrimPrefix(number, "ПОДР-")
		m[code] = id
	}
	return m, nil
}

// isPersonName checks if the value looks like a person name (has spaces, Cyrillic characters)
func isPersonName(s string) bool {
	if len(s) < 3 {
		return false
	}
	if !strings.Contains(s, " ") {
		return false
	}
	for _, r := range s {
		if r >= 0x0400 && r <= 0x04FF {
			return true
		}
	}
	return false
}

// parseCardNumber parses card number from string, handling both int and float formats
func parseCardNumber(s string) (int, error) {
	if n, err := strconv.Atoi(s); err == nil {
		return n, nil
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int(f), nil
	}
	return 0, strconv.ErrSyntax
}

// isCardNumber checks if the string looks like a card number (integer or float that rounds to integer)
func isCardNumber(s string) bool {
	if n, err := strconv.Atoi(s); err == nil {
		return n > 0 && n < 10000
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		n := int(f)
		return n > 0 && n < 10000 && float64(n) == f
	}
	return false
}
