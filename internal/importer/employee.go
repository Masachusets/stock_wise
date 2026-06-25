package importer

import (
	"context"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

func (imp *Importer) importEmployees(ctx context.Context) error {
	path := imp.excelPath("UMS.xlsx")
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.GetRows("КУ Ф-111")
	if err != nil {
		return err
	}

	type emp struct {
		cardNumber int
		fullName   string
		isActive   bool
	}
	var items []emp

	for i, row := range rows {
		if i == 0 {
			continue // header
		}
		if len(row) < 4 {
			continue
		}
		cardStr := strings.TrimSpace(row[1])
		name := strings.TrimSpace(row[3])
		note := ""
		if len(row) > 5 {
			note = strings.TrimSpace(row[5])
		}
		if cardStr == "" || name == "" {
			continue
		}
		card, err := strconv.Atoi(cardStr)
		if err != nil {
			continue
		}
		name = stripRank(name)
		isActive := !strings.EqualFold(note, "СПИСАНА")
		items = append(items, emp{cardNumber: card, fullName: name, isActive: isActive})
	}

	log.Printf("employees: найдено %d записей", len(items))

	tx, err := imp.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		_, err := tx.Exec(ctx,
			`INSERT INTO cards (number, full_name, is_active) VALUES ($1, $2, $3) ON CONFLICT (number) DO NOTHING`,
			item.cardNumber, item.fullName, item.isActive,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// stripRank удаляет воинское звание (или любые другие служебные слова) из начала строки,
// оставляя только фамилию, имя и отчество (ФИО). Функция ищет первую заглавную букву
// и возвращает подстроку, начиная с неё, предварительно обрезая пробелы.
// Если заглавная буква не найдена, возвращается исходная строка без изменений.
func stripRank(name string) string {
    s := strings.TrimSpace(name)
    if s == "" {
        return ""
    }
    for i, r := range s {
        if unicode.IsUpper(r) {
            return strings.TrimSpace(s[i:])
        }
    }
    return s
}