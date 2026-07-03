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
		a.target_type, c.full_name, d.name
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	LEFT JOIN equipments_assignments a ON a.equipment_id = e.id AND a.is_active = TRUE
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN departments d ON a.department_code = d.code
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
		var targetType, empFullName, deptName interface{}
		if err := rows.Scan(
			&invNum, &serialNum,
			&nomCode, &nomName,
			&modelName, &mfgDate, &arrDate,
			&status, &formNum, &loc, &notes,
			&targetType, &empFullName, &deptName,
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
				TargetType: fmt.Sprintf("%v", targetType),
				FullName:   strPtr(empFullName),
				DeptName:   strPtr(deptName),
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
	var cardNum, deptName, fullName, opComment interface{}
	err = r.db.QueryRow(ctx, `SELECT
		a.target_type,
		a.card_number,
		c.full_name,
		d.name,
		a.operator_comment
	FROM equipments_assignments a
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN departments d ON a.department_code = d.code
	WHERE a.equipment_id = $1 AND a.is_active = TRUE`, eqID).Scan(
		&ai.TargetType,
		&cardNum,
		&fullName,
		&deptName,
		&opComment,
	)
	if err == nil {
		if cardNum != nil {
			var cn int32
			fmt.Sscanf(fmt.Sprintf("%v", cardNum), "%d", &cn)
			ai.CardNumber = &cn
		}
		ai.FullName = strPtr(fullName)
		ai.DeptName = strPtr(deptName)
		ai.OperatorComment = strPtr(opComment)
		e.Assignment = ai
	}

	return e, nil
}

func (r *postgresRepository) Create(ctx context.Context, eq *Equipment) error {
	tag, err := r.db.Exec(ctx, `INSERT INTO equipments
		(inventory_number, serial_number, nomenclature_id, model_name, manufacture_date, arrival_date, status, form_number, location, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		eq.InventoryNumber, eq.SerialNumber, eq.NomenclatureID, eq.ModelName,
		eq.ManufactureDate, eq.ArrivalDate, eq.Status, eq.FormNumber, eq.Location, eq.Notes)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("оборудование с инвентарным номером %s уже существует", eq.InventoryNumber)
	}
	return nil
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

func (r *postgresRepository) ListForWeb(ctx context.Context, filter *ListFilter) ([]*EquipmentListItem, error) {
	query := `SELECT
		e.inventory_number,
		COALESCE(e.model_name, ''),
		e.status,
		e.location,
		n.code, n.name,
		a.target_type, c.full_name, d.name
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	LEFT JOIN equipments_assignments a ON a.equipment_id = e.id AND a.is_active = TRUE
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN departments d ON a.department_code = d.code
	WHERE e.deleted_at IS NULL`

	args := []any{}
	argIdx := 1

	if filter != nil {
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
	}

	query += ` ORDER BY e.inventory_number`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*EquipmentListItem
	for rows.Next() {
		item := &EquipmentListItem{}
		var nomCode, nomName, targetType, empFullName, deptName interface{}
		if err := rows.Scan(
			&item.InventoryNumber, &item.ModelName, &item.Status, &item.Location,
			&nomCode, &nomName, &targetType, &empFullName, &deptName,
		); err != nil {
			return nil, err
		}
		if nomCode != nil {
			item.Nomenclature = &NomenclatureInfo{
				Code: fmt.Sprintf("%v", nomCode),
				Name: fmt.Sprintf("%v", nomName),
			}
		}
		if targetType != nil {
			item.Assignment = &AssignmentInfo{
				TargetType: fmt.Sprintf("%v", targetType),
				FullName:   strPtr(empFullName),
				DeptName:   strPtr(deptName),
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepository) GetForWeb(ctx context.Context, inventoryNumber string) (*EquipmentDetail, error) {
	e := &EquipmentDetail{}
	var nomCode, nomName interface{}
	err := r.db.QueryRow(ctx, `SELECT
		e.inventory_number, e.serial_number,
		COALESCE(e.model_name, ''), e.manufacture_date::text, e.arrival_date::text,
		e.status, e.form_number, e.location, e.notes,
		n.code, n.name
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE e.inventory_number = $1 AND e.deleted_at IS NULL`, inventoryNumber).Scan(
		&e.InventoryNumber, &e.SerialNumber,
		&e.ModelName, &e.ManufactureDate, &e.ArrivalDate,
		&e.Status, &e.FormNumber, &e.Location, &e.Notes,
		&nomCode, &nomName,
	)
	if err != nil {
		return nil, err
	}
	if nomCode != nil {
		e.Nomenclature = &NomenclatureInfo{
			Code: fmt.Sprintf("%v", nomCode),
			Name: fmt.Sprintf("%v", nomName),
		}
	}

	// Получить закрепление
	var eqID int32
	err = r.db.QueryRow(ctx, "SELECT id FROM equipments WHERE inventory_number = $1", inventoryNumber).Scan(&eqID)
	if err == nil {
		ai := &AssignmentInfo{}
		var cardNum, deptName, fullName, opComment interface{}
		err = r.db.QueryRow(ctx, `SELECT
			a.target_type, a.card_number, c.full_name, d.name, a.operator_comment
		FROM equipments_assignments a
		LEFT JOIN cards c ON a.card_number = c.number
		LEFT JOIN departments d ON a.department_code = d.code
		WHERE a.equipment_id = $1 AND a.is_active = TRUE`, eqID).Scan(
			&ai.TargetType, &cardNum, &fullName, &deptName, &opComment,
		)
		if err == nil {
			if cardNum != nil {
				var cn int32
				fmt.Sscanf(fmt.Sprintf("%v", cardNum), "%d", &cn)
				ai.CardNumber = &cn
			}
			ai.FullName = strPtr(fullName)
			ai.DeptName = strPtr(deptName)
			ai.OperatorComment = strPtr(opComment)
			e.Assignment = ai
		}
	}

	return e, nil
}

func (r *postgresRepository) ListNomenclatures(ctx context.Context) ([]*NomenclatureOption, error) {
	rows, err := r.db.Query(ctx, "SELECT id, code, name FROM nomenclatures ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*NomenclatureOption
	for rows.Next() {
		n := &NomenclatureOption{}
		if err := rows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, nil
}

func (r *postgresRepository) CreateWithAssignment(ctx context.Context, eq *Equipment, departmentCode int) error {
	if err := r.Create(ctx, eq); err != nil {
		return err
	}

	var eqID int32
	err := r.db.QueryRow(ctx, "SELECT id FROM equipments WHERE inventory_number = $1", eq.InventoryNumber).Scan(&eqID)
	if err != nil {
		return nil
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO equipments_assignments (equipment_id, target_type, department_code)
		 VALUES ($1, 'warehouse', $2)`, eqID, departmentCode)
	return err
}

func (r *postgresRepository) ListDeleted(ctx context.Context) ([]*EquipmentDeletedItem, error) {
	rows, err := r.db.Query(ctx, `SELECT
		e.inventory_number, COALESCE(e.model_name, ''),
		n.code, n.name, e.status, e.deleted_at::text
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE e.deleted_at IS NOT NULL
	ORDER BY e.deleted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*EquipmentDeletedItem
	for rows.Next() {
		item := &EquipmentDeletedItem{}
		if err := rows.Scan(&item.InventoryNumber, &item.ModelName, &item.NomCode, &item.NomName, &item.Status, &item.DeletedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
