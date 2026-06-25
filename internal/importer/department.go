package importer

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type departmentItem struct {
	code int
	name string
	typ  string
}

func (imp *Importer) importDepartments(ctx context.Context) error {
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

	// Collect unique department names from column M
	deptSet := make(map[string]bool)
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 13 {
			continue
		}
		colM := strings.TrimSpace(row[12])
		if colM == "" {
			continue
		}
		if colM == "!" {
			colM = "СКЛАД"
		}
		if isDepartment(colM) {
			deptSet[colM] = true
		}
	}

	// Parse each department name into type + name
	var items []departmentItem
	typeCounters := map[string]int{
		"warehouse": 0,
		"upogg":     0,
		"opk":       0,
		"pogk":      0,
		"pogz":      0,
	}
	typeDigits := map[string]string{
		"warehouse": "1",
		"upogg":     "2",
		"opk":       "3",
		"pogk":      "4",
		"pogz":      "5",
	}

	for raw := range deptSet {
		typ, name := parseDeptTypeAndName(raw)
		typeCounters[typ]++
		codeStr := typeDigits[typ] + padNum(typeCounters[typ])
		code, _ := strconv.Atoi(codeStr)
		items = append(items, departmentItem{code: code, name: name, typ: typ})
	}

	// Sort by code for deterministic output
	sort.Slice(items, func(i, j int) bool {
		return items[i].code < items[j].code
	})

	log.Printf("departments: найдено %d уникальных подразделений", len(items))

	// Ensure default departments exist
	defaults := []departmentItem{
		{code: 100, name: "СКЛАД", typ: "warehouse"},
		{code: 200, name: "УПОГГ", typ: "upogg"},
	}
	items = append(defaults, items...)

	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO departments (code, type, name) VALUES ($1, $2, $3) ON CONFLICT (name) DO NOTHING`,
			item.code, item.typ, item.name,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (imp *Importer) createDepartmentWaybills(ctx context.Context) error {
	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT code, name FROM departments`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type dept struct {
		code int
		name string
	}
	var depts []dept
	for rows.Next() {
		var d dept
		if err := rows.Scan(&d.code, &d.name); err != nil {
			return err
		}
		depts = append(depts, d)
	}

	for _, d := range depts {
		wbNumber := "ПОДР-" + strconv.Itoa(d.code)
		_, err := tx.Exec(ctx,
			`INSERT INTO waybills (number, issue_date, from_dept, to_dept, status)
			 VALUES ($1, CURRENT_DATE, $2, $2, 'archived')
			 ON CONFLICT (number) DO NOTHING`,
			wbNumber, d.code,
		)
		if err != nil {
			return fmt.Errorf("creating waybill for dept %d: %w", d.code, err)
		}
	}

	log.Printf("waybills: создано %d накладных для подразделений", len(depts))
	return tx.Commit(ctx)
}

func parseDeptTypeAndName(raw string) (typ string, name string) {
	lower := strings.ToLower(raw)

	// Check for type keywords (order matters: longer first)
	typeKeywords := []struct {
		keyword string
		typ     string
	}{
		{"склад", "warehouse"},
		{"упогг", "upogg"},
		{"погк", "pogk"},
		{"погз", "pogz"},
		{"опк", "opk"},
	}

	for _, tk := range typeKeywords {
		if idx := strings.Index(lower, tk.keyword); idx >= 0 {
			// Remove the keyword and any separators around it
			rest := raw
			// Remove keyword
			rest = rest[:idx] + rest[idx+len(tk.keyword):]
			// Remove leading/trailing separators
			rest = strings.Trim(rest, "/ ")
			if rest == "" {
				rest = raw
			}
			return tk.typ, rest
		}
	}

	// Default: no type keyword found, treat as warehouse
	return "warehouse", raw
}

func padNum(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}

func isDepartment(s string) bool {
	lower := strings.ToLower(s)

	// Known department patterns
	deptPatterns := []string{
		"склад", "опк", "погк", "погз", "упогг",
		"аэропорт", "пункт", "база", "центр",
		"ж/д", "клас",
	}

	for _, kw := range deptPatterns {
		if strings.Contains(lower, kw) {
			return true
		}
	}

	// Single-word short strings (like location codes: 219, Тир)
	if len(s) <= 10 && !strings.Contains(s, " ") && !strings.Contains(s, ".") {
		return true
	}

	return false
}
