package waybills

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context) ([]*Waybill, error) {
	rows, err := r.db.Query(ctx, "SELECT id, number, issue_date::text, from_dept, to_dept, status FROM waybills WHERE deleted_at IS NULL ORDER BY issue_date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waybills []*Waybill
	for rows.Next() {
		w := &Waybill{}
		if err := rows.Scan(&w.ID, &w.Number, &w.IssueDate, &w.FromDept, &w.ToDept, &w.Status); err != nil {
			return nil, err
		}
		waybills = append(waybills, w)
	}
	return waybills, nil
}

func (r *postgresRepository) Get(ctx context.Context, id int32) (*Waybill, error) {
	w := &Waybill{}
	err := r.db.QueryRow(ctx, "SELECT id, number, issue_date::text, from_dept, to_dept, status FROM waybills WHERE id = $1", id).Scan(&w.ID, &w.Number, &w.IssueDate, &w.FromDept, &w.ToDept, &w.Status)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, "SELECT waybill_id, equipment_id FROM waybills_equipments WHERE waybill_id = $1", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		we := &WaybillsEquipment{}
		if err := rows.Scan(&we.WaybillID, &we.EquipmentID); err != nil {
			return nil, err
		}
		w.Items = append(w.Items, we)
	}
	return w, nil
}

func (r *postgresRepository) Create(ctx context.Context, wb *Waybill) error {
	var id int32
	err := r.db.QueryRow(ctx,
		"INSERT INTO waybills (number, issue_date, from_dept, to_dept) VALUES ($1, $2, $3, $4) RETURNING id",
		wb.Number, wb.IssueDate, wb.FromDept, wb.ToDept,
	).Scan(&id)
	if err != nil {
		return err
	}
	wb.ID = id

	for _, item := range wb.Items {
		_, err = r.db.Exec(ctx, "INSERT INTO waybills_equipments (waybill_id, equipment_id) VALUES ($1, $2)", id, item.EquipmentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *postgresRepository) GetEquipmentIDs(ctx context.Context, waybillID int32) ([]int32, error) {
	rows, err := r.db.Query(ctx, "SELECT equipment_id FROM waybills_equipments WHERE waybill_id = $1", waybillID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *postgresRepository) UpdateStatus(ctx context.Context, id int32, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE waybills SET status = $1 WHERE id = $2", status, id)
	return err
}

func (r *postgresRepository) CreateAssignment(ctx context.Context, equipmentID int32, departmentCode int32) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO equipments_assignments (equipment_id, target_type, department_code)
		 VALUES ($1, 'department', $2)`, equipmentID, departmentCode)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, id int32) error {
	_, err := r.db.Exec(ctx, "DELETE FROM waybills WHERE id = $1", id)
	return err
}

func (r *postgresRepository) GetStatus(ctx context.Context, id int32) (string, error) {
	var status string
	err := r.db.QueryRow(ctx, "SELECT status FROM waybills WHERE id = $1", id).Scan(&status)
	return status, err
}

func (r *postgresRepository) ListForWeb(ctx context.Context) ([]*WaybillListItem, error) {
	rows, err := r.db.Query(ctx, `SELECT w.id, w.number, w.issue_date::text, w.status,
		fd.name, td.name
	FROM waybills w
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE w.deleted_at IS NULL
	ORDER BY w.issue_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*WaybillListItem
	for rows.Next() {
		w := &WaybillListItem{}
		if err := rows.Scan(&w.ID, &w.Number, &w.IssueDate, &w.Status, &w.FromName, &w.ToName); err != nil {
			return nil, err
		}
		items = append(items, w)
	}
	return items, nil
}

func (r *postgresRepository) GetForWeb(ctx context.Context, id int32) (*WaybillDetail, error) {
	w := &WaybillDetail{}
	err := r.db.QueryRow(ctx, `SELECT w.number, w.issue_date::text, w.status,
		fd.name, td.name
	FROM waybills w
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE w.id = $1`, id).Scan(&w.Number, &w.IssueDate, &w.Status, &w.FromName, &w.ToName)
	if err != nil {
		return nil, err
	}

	eqRows, err := r.db.Query(ctx, `SELECT DISTINCT e.inventory_number, COALESCE(e.model_name, ''), n.name
	FROM equipments_assignments a
	JOIN equipments e ON a.equipment_id = e.id
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE a.department_code = (SELECT to_dept FROM waybills WHERE id = $1) AND a.is_active = TRUE`, id)
	if err != nil {
		return nil, err
	}
	defer eqRows.Close()

	for eqRows.Next() {
		var item WaybillEquipmentItem
		var modelName, nomenclature interface{}
		if err := eqRows.Scan(&item.InventoryNumber, &modelName, &nomenclature); err != nil {
			return nil, err
		}
		if modelName != nil {
			item.ModelName = fmt.Sprintf("%v", modelName)
		}
		if nomenclature != nil {
			item.Nomenclature = fmt.Sprintf("%v", nomenclature)
		}
		w.Equipments = append(w.Equipments, item)
	}
	return w, nil
}
