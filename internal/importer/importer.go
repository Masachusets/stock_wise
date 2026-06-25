package importer

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Importer struct {
	db  *pgxpool.Pool
	dir string
}

func New(db *pgxpool.Pool, dir string) *Importer {
	return &Importer{db: db, dir: dir}
}

func (imp *Importer) Run(ctx context.Context) error {
	log.Println("=== Начало импорта ===")

	// Clear existing data for fresh import
	if err := imp.clearTables(ctx); err != nil {
		return fmt.Errorf("clearing tables: %w", err)
	}

	if err := imp.importNomenclatures(ctx); err != nil {
		return fmt.Errorf("nomenclatures: %w", err)
	}

	if err := imp.importEmployees(ctx); err != nil {
		return fmt.Errorf("employees: %w", err)
	}

	if err := imp.importDepartments(ctx); err != nil {
		return fmt.Errorf("departments: %w", err)
	}

	if err := imp.createDepartmentWaybills(ctx); err != nil {
		return fmt.Errorf("department waybills: %w", err)
	}

	if err := imp.importEquipments(ctx); err != nil {
		return fmt.Errorf("equipments: %w", err)
	}

	if err := imp.importAssignments(ctx); err != nil {
		return fmt.Errorf("assignments: %w", err)
	}

	log.Println("=== Импорт завершён ===")
	return nil
}

func (imp *Importer) excelPath(name string) string {
	return filepath.Join(imp.dir, name)
}

func cellString(v interface{}) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", v))
}

func statusMap(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case strings.HasPrefix(s, "эксп(инт)"):
		return "exp_int"
	case strings.HasPrefix(s, "эксп(сп)"):
		return "exp_sp"
	case strings.Contains(s, "неиспр"):
		return "broken"
	case strings.HasPrefix(s, "эксп"):
		return "exp"
	case strings.Contains(s, "списать"):
		return "written_off"
	default:
		return "exp"
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (imp *Importer) clearTables(ctx context.Context) error {
	tables := []string{
		"equipments_assignments",
		"waybills_equipments",
		"equipments",
		"waybills",
		"departments",
		"cards",
		"nomenclatures",
	}
	for _, t := range tables {
		if _, err := imp.db.Exec(ctx, "TRUNCATE TABLE "+t+" CASCADE"); err != nil {
			return err
		}
	}
	log.Println("tables cleared")
	return nil
}
