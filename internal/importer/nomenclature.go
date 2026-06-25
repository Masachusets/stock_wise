package importer

import (
	"context"
	"log"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

var nomenclatureRe = regexp.MustCompile(`^\d{2}/\d+\.\d+\.\d+\.\d+\.\d+/\d{4}$|^\d{2}\.\d+\.\d+\.\d+\.\d+/\d{4}$`)

func (imp *Importer) importNomenclatures(ctx context.Context) error {
	path := imp.excelPath("Номенклатурная таблица 01.2026.xlsx")
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	type nom struct {
		code string
		name string
	}
	seen := make(map[string]bool)
	var items []nom

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			continue
		}
		for _, row := range rows {
			if len(row) < 2 {
				continue
			}
			code := strings.TrimSpace(row[0])
			name := strings.TrimSpace(row[1])
			if code == "" || name == "" {
				continue
			}
			if !nomenclatureRe.MatchString(code) {
				continue
			}
			if seen[code] {
				continue
			}
			seen[code] = true
			items = append(items, nom{code: code, name: name})
		}
	}

	log.Printf("nomenclatures: найдено %d записей", len(items))

	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO nomenclatures (code, name) VALUES ($1, $2) ON CONFLICT (code) DO NOTHING`,
			item.code, item.name,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
