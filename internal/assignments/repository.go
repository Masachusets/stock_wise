package assignments

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context, equipmentID *int32, isActive *bool) ([]*Assignment, error) {
	query := `SELECT
		a.target_type,
		a.card_number,
		c.full_name,
		w.number,
		w.issue_date::text,
		fd.name,
		td.name,
		a.assigned_at::text,
		a.unassigned_at::text,
		a.operator_comment
	FROM equipments_assignments a
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN waybills w ON a.waybill_id = w.id
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE 1=1`

	args := []any{}
	argIdx := 1

	if equipmentID != nil {
		query += ` AND a.equipment_id = $` + strconv.Itoa(argIdx)
		args = append(args, *equipmentID)
		argIdx++
	}
	if isActive != nil {
		query += ` AND a.is_active = $` + strconv.Itoa(argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	query += ` ORDER BY a.assigned_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*Assignment
	for rows.Next() {
		a := &Assignment{}
		if err := rows.Scan(
			&a.TargetType,
			&a.CardNumber,
			&a.FullName,
			&a.WaybillNumber,
			&a.WaybillDate,
			&a.FromDeptName,
			&a.ToDeptName,
			&a.AssignedAt,
			&a.UnassignedAt,
			&a.OperatorComment,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (r *postgresRepository) Get(ctx context.Context, id int32) (*Assignment, error) {
	a := &Assignment{}
	err := r.db.QueryRow(ctx, `SELECT
		a.target_type,
		a.card_number,
		c.full_name,
		w.number,
		w.issue_date::text,
		fd.name,
		td.name,
		a.assigned_at::text,
		a.unassigned_at::text,
		a.operator_comment
	FROM equipments_assignments a
	LEFT JOIN cards c ON a.card_number = c.number
	LEFT JOIN waybills w ON a.waybill_id = w.id
	LEFT JOIN departments fd ON w.from_dept = fd.code
	LEFT JOIN departments td ON w.to_dept = td.code
	WHERE a.id = $1`, id).Scan(
		&a.TargetType,
		&a.CardNumber,
		&a.FullName,
		&a.WaybillNumber,
		&a.WaybillDate,
		&a.FromDeptName,
		&a.ToDeptName,
		&a.AssignedAt,
		&a.UnassignedAt,
		&a.OperatorComment,
	)
	return a, err
}
