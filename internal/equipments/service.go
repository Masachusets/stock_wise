package equipments

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	gen "github.com/Masachusets/stock_wise/gen/equipments"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func strPtr(v interface{}) *string {
	if v == nil {
		return nil
	}
	s := fmt.Sprintf("%v", v)
	return &s
}

func (s *service) List(ctx context.Context, p *gen.ListPayload) (res *gen.EquipmentList, err error) {
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

	if p.Status != nil {
		query += ` AND e.status = $` + strconv.Itoa(argIdx)
		args = append(args, *p.Status)
		argIdx++
	}
	if p.NomenclatureID != nil {
		query += ` AND e.nomenclature_id = $` + strconv.Itoa(argIdx)
		args = append(args, *p.NomenclatureID)
		argIdx++
	}
	if p.Location != nil {
		query += ` AND e.location = $` + strconv.Itoa(argIdx)
		args = append(args, *p.Location)
		argIdx++
	}
	if p.Search != nil {
		query += ` AND to_tsvector('russian', e.model_name || ' ' || COALESCE(e.serial_number, '')) @@ plainto_tsquery('russian', $` + strconv.Itoa(argIdx) + `)`
		args = append(args, *p.Search)
		argIdx++
	}

	query += ` ORDER BY e.inventory_number`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipments []*gen.Equipment
	for rows.Next() {
		e := &gen.Equipment{}
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
			n := &gen.Nomenclature{
				Code: fmt.Sprintf("%v", nomCode),
				Name: fmt.Sprintf("%v", nomName),
			}
			e.Nomenclature = n
		}
		if targetType != nil {
			ai := &gen.AssignmentInfo{
				TargetType: fmt.Sprintf("%v", targetType),
				FullName:   strPtr(empFullName),
				WaybillNumber: strPtr(wbNumber),
				ToDeptName: strPtr(deptName),
			}
			e.Assignment = ai
		}
		equipments = append(equipments, e)
	}
	return &gen.EquipmentList{Equipments: equipments}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Equipment, err error) {
	res = &gen.Equipment{}
	var invNum, status, modelName, serialNum, nomCode, nomName, mfgDate, arrDate, formNum, loc, notes interface{}
	err = s.db.QueryRow(ctx, `SELECT
		e.inventory_number, e.serial_number,
		n.code, n.name,
		COALESCE(e.model_name, ''), e.manufacture_date::text, e.arrival_date::text,
		e.status, e.form_number, e.location, e.notes
	FROM equipments e
	LEFT JOIN nomenclatures n ON e.nomenclature_id = n.id
	WHERE e.inventory_number = $1 AND e.deleted_at IS NULL`, p.InventoryNumber).Scan(
		&invNum, &serialNum,
		&nomCode, &nomName,
		&modelName, &mfgDate, &arrDate,
		&status, &formNum, &loc, &notes,
	)
	if err != nil {
		return nil, err
	}
	res.InventoryNumber = fmt.Sprintf("%v", invNum)
	res.Status = fmt.Sprintf("%v", status)
	res.ModelName = fmt.Sprintf("%v", modelName)
	res.SerialNumber = strPtr(serialNum)
	res.ManufactureDate = strPtr(mfgDate)
	res.ArrivalDate = strPtr(arrDate)
	res.FormNumber = strPtr(formNum)
	res.Location = strPtr(loc)
	res.Notes = strPtr(notes)
	if nomCode != nil {
		n := &gen.Nomenclature{
			Code: fmt.Sprintf("%v", nomCode),
			Name: fmt.Sprintf("%v", nomName),
		}
		res.Nomenclature = n
	}

	var eqID int32
	err = s.db.QueryRow(ctx, "SELECT id FROM equipments WHERE inventory_number = $1", p.InventoryNumber).Scan(&eqID)
	if err != nil {
		return res, nil
	}

	ai := &gen.AssignmentInfo{}
	var cardNum, wbNum, wbDate, fdName, tdName, fullName, opComment interface{}
	err = s.db.QueryRow(ctx, `SELECT
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
			n, _ := fmt.Sprintf("%v", cardNum), error(nil)
			_ = n
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
		res.Assignment = ai
	}

	return
}

func (s *service) Create(ctx context.Context, p *gen.CreateEquipmentPayload) (res *gen.Equipment, err error) {
	_, err = s.db.Exec(ctx, `INSERT INTO equipments
		(inventory_number, serial_number, nomenclature_id, model_name, manufacture_date, arrival_date, status, form_number, location, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (inventory_number) DO NOTHING`,
		p.InventoryNumber, p.SerialNumber, p.NomenclatureID, p.ModelName,
		p.ManufactureDate, p.ArrivalDate, p.Status, p.FormNumber, p.Location, p.Notes)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{InventoryNumber: p.InventoryNumber})
}

func (s *service) Update(ctx context.Context, p *gen.UpdateEquipmentPayload) (res *gen.Equipment, err error) {
	sets := []string{}
	args := []any{}
	argIdx := 1

	if p.SerialNumber != nil {
		sets = append(sets, fmt.Sprintf("serial_number = $%d", argIdx))
		args = append(args, *p.SerialNumber)
		argIdx++
	}
	if p.NomenclatureID != nil {
		sets = append(sets, fmt.Sprintf("nomenclature_id = $%d", argIdx))
		args = append(args, *p.NomenclatureID)
		argIdx++
	}
	if p.ModelName != nil {
		sets = append(sets, fmt.Sprintf("model_name = $%d", argIdx))
		args = append(args, *p.ModelName)
		argIdx++
	}
	if p.ManufactureDate != nil {
		sets = append(sets, fmt.Sprintf("manufacture_date = $%d", argIdx))
		args = append(args, *p.ManufactureDate)
		argIdx++
	}
	if p.ArrivalDate != nil {
		sets = append(sets, fmt.Sprintf("arrival_date = $%d", argIdx))
		args = append(args, *p.ArrivalDate)
		argIdx++
	}
	if p.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *p.Status)
		argIdx++
	}
	if p.FormNumber != nil {
		sets = append(sets, fmt.Sprintf("form_number = $%d", argIdx))
		args = append(args, *p.FormNumber)
		argIdx++
	}
	if p.Location != nil {
		sets = append(sets, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *p.Location)
		argIdx++
	}
	if p.Notes != nil {
		sets = append(sets, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *p.Notes)
		argIdx++
	}

	if len(sets) == 0 {
		return s.Get(ctx, &gen.GetPayload{InventoryNumber: p.InventoryNumber})
	}

	args = append(args, p.InventoryNumber)
	query := fmt.Sprintf("UPDATE equipments SET %s WHERE inventory_number = $%d", strings.Join(sets, ", "), argIdx)
	_, err = s.db.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{InventoryNumber: p.InventoryNumber})
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	_, err := s.db.Exec(ctx, "UPDATE equipments SET deleted_at = NOW() WHERE inventory_number = $1 AND deleted_at IS NULL", p.InventoryNumber)
	return err
}
