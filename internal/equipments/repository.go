package equipments

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func strPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	s := fmt.Sprintf("%v", v)
	return &s
}

func (r *postgresRepository) List(ctx context.Context, filter *ListFilter) ([]*Equipment, error) {
	query := `SELECT
		e.inventory_number, e.serial_number,
		n.code, n.name,
		COALESCE(e.model_name, ''), e.manufacture_date::text, e.arrival_date::text,
		e.status, e.form_number, e.location, e.notes,
		a.target_type, c.full_name, w.number, td.name
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	LEFT JOIN equipments_assignments a ON a.equipment_id = e.id AND a.is_active = TRUE
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN waybills w ON a.waybill_id = w.id
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE e.deleted_at IS NULL`

	args := []any{}
	argIdx := 1

	if filter.Status != nil {
		query += ` AND e.status = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.NomenclatureID != nil {
		query += ` AND e.nomenclature_id = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.NomenclatureID)
		argIdx++
	}
	if filter.Location != nil {
		query += ` AND e.location = $` + strconv.Itoa(argIdx)
		args = append(args, *filter.Location)
		argIdx++
	}
	if filter.Search != nil {
		query += ` AND to_tsvector('russian', e.model_name || ' ' || COALESCE(e.serial_number, '')) @@ plainto_tsquery('russian', $` + strconv.Itoa(argIdx) + `)`
		args = append(args, *filter.Search)
		argIdx++
	}

	query += ` ORDER BY e.inventory_number`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipments []*Equipment
	for rows.Next() {
		e := &Equipment{}
		var invNum, status, modelName, serialNum, nomCode, nomName, mfgDate, arrDate, formNum, loc, notes interface{}
		var targetType, empFullName, wbNumber, deptName interface{}
		if err := rows.Scan(
			&invNum, &serialNum,
			&nomCode, &nomName,
			&modelName, &mfgDate, &arrDate,
			&status, &formNum, &loc, &notes,
			&targetType, &empFullName, &wbNumber, &deptName,
		); err != nil {
			return nil, err
		}
		e.InventoryNumber = fmt.Sprintf("%v", invNum)
		e.Status = fmt.Sprintf("%v", status)
		e.ModelName = fmt.Sprintf("%v", modelName)
		e.SerialNumber = strPtr(serialNum)
		e.ManufactureDate = strPtr(mfgDate)
		e.ArrivalDate = strPtr(arrDate)
		e.FormNumber = strPtr(formNum)
		e.Location = strPtr(loc)
		e.Notes = strPtr(notes)
		if nomCode != nil {
			e.Nomenclature = &NomenclatureInfo{
				Code: fmt.Sprintf("%v", nomCode),
				Name: fmt.Sprintf("%v", nomName),
			}
		}
		if targetType != nil {
			e.Assignment = &AssignmentInfo{
				TargetType:   fmt.Sprintf("%v", targetType),
				FullName:     strPtr(empFullName),
				WaybillNumber: strPtr(wbNumber),
				ToDeptName:   strPtr(deptName),
			}
		}
		equipments = append(equipments, e)
	}
	return equipments, nil
}

func (r *postgresRepository) Get(ctx context.Context, inventoryNumber string) (*Equipment, error) {
	e := &Equipment{}
	var invNum, status, modelName, serialNum, nomCode, nomName, mfgDate, arrDate, formNum, loc, notes interface{}
	err := r.db.QueryRow(ctx, `SELECT
		e.inventory_number, e.serial_number,
		n.code, n.name,
		COALESCE(e.model_name, ''), e.manufacture_date::text, e.arrival_date::text,
		e.status, e.form_number, e.location, e.notes
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE e.inventory_number = $1 AND e.deleted_at IS NULL`, inventoryNumber).Scan(
		&invNum, &serialNum,
		&nomCode, &nomName,
		&modelName, &mfgDate, &arrDate,
		&status, &formNum, &loc, &notes,
	)
	if err != nil {
		return nil, err
	}
	e.InventoryNumber = fmt.Sprintf("%v", invNum)
	e.Status = fmt.Sprintf("%v", status)
	e.ModelName = fmt.Sprintf("%v", modelName)
	e.SerialNumber = strPtr(serialNum)
	e.ManufactureDate = strPtr(mfgDate)
	e.ArrivalDate = strPtr(arrDate)
	e.FormNumber = strPtr(formNum)
	e.Location = strPtr(loc)
	e.Notes = strPtr(notes)
	if nomCode != nil {
		e.Nomenclature = &NomenclatureInfo{
			Code: fmt.Sprintf("%v", nomCode),
			Name: fmt.Sprintf("%v", nomName),
		}
	}

	// Получить закрепление
	var eqID int32
	err = r.db.QueryRow(ctx, "SELECT id FROM equipments WHERE inventory_number = $1", inventoryNumber).Scan(&eqID)
	if err != nil {
		return e, nil
	}

	ai := &AssignmentInfo{}
	var cardNum, wbNum, wbDate, fdName, tdName, fullName, opComment interface{}
	err = r.db.QueryRow(ctx, `SELECT
		a.target_type,
		a.card_number,
		c.full_name,
		w.number,
		w.issue_date::text,
		fd.name,
		td.name,
		a.operator_comment
	FROM equipments_assignments a
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN waybills w ON a.waybill_id = w.id
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE a.equipment_id = $1 AND a.is_active = TRUE`, eqID).Scan(
		&ai.TargetType,
		&cardNum,
		&fullName,
		&wbNum,
		&wbDate,
		&fdName,
		&tdName,
		&opComment,
	)
	if err == nil {
		if cardNum != nil {
			var cn int32
			fmt.Sscanf(fmt.Sprintf("%v", cardNum), "%d", &cn)
			ai.CardNumber = &cn
		}
		ai.FullName = strPtr(fullName)
		ai.WaybillNumber = strPtr(wbNum)
		ai.WaybillDate = strPtr(wbDate)
		ai.FromDeptName = strPtr(fdName)
		ai.ToDeptName = strPtr(tdName)
		ai.OperatorComment = strPtr(opComment)
		e.Assignment = ai
	}

	return e, nil
}

func (r *postgresRepository) Create(ctx context.Context, eq *Equipment) error {
	_, err := r.db.Exec(ctx, `INSERT INTO equipments
		(inventory_number, serial_number, nomenclature_id, model_name, manufacture_date, arrival_date, status, form_number, location, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (inventory_number) DO NOTHING`,
		eq.InventoryNumber, eq.SerialNumber, nil, eq.ModelName,
		eq.ManufactureDate, eq.ArrivalDate, eq.Status, eq.FormNumber, eq.Location, eq.Notes)
	return err
}

func (r *postgresRepository) Update(ctx context.Context, eq *Equipment) error {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if eq.SerialNumber != nil {
		sets = append(sets, fmt.Sprintf("serial_number = $%d", argIdx))
		args = append(args, *eq.SerialNumber)
		argIdx++
	}
	if eq.ModelName != "" {
		sets = append(sets, fmt.Sprintf("model_name = $%d", argIdx))
		args = append(args, eq.ModelName)
		argIdx++
	}
	if eq.ManufactureDate != nil {
		sets = append(sets, fmt.Sprintf("manufacture_date = $%d", argIdx))
		args = append(args, *eq.ManufactureDate)
		argIdx++
	}
	if eq.ArrivalDate != nil {
		sets = append(sets, fmt.Sprintf("arrival_date = $%d", argIdx))
		args = append(args, *eq.ArrivalDate)
		argIdx++
	}
	if eq.Status != "" {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, eq.Status)
		argIdx++
	}
	if eq.FormNumber != nil {
		sets = append(sets, fmt.Sprintf("form_number = $%d", argIdx))
		args = append(args, *eq.FormNumber)
		argIdx++
	}
	if eq.Location != nil {
		sets = append(sets, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *eq.Location)
		argIdx++
	}
	if eq.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *eq.Notes)
		argIdx++
	}

	if len(sets) == 0 {
		return nil
	}

	args = append(args, eq.InventoryNumber)
	query := fmt.Sprintf("UPDATE equipments SET %s WHERE inventory_number = $%d", strings.Join(sets, ", "), argIdx)
	_, err := r.db.Exec(ctx, query, args...)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, inventoryNumber string) error {
	_, err := r.db.Exec(ctx, "UPDATE equipments SET deleted_at = NOW() WHERE inventory_number = $1 AND deleted_at IS NULL", inventoryNumber)
	return err
}
