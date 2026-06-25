package assignments

import (
	"context"
	"strconv"

	gen "github.com/Masachusets/stock_wise/gen/assignments"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context, p *gen.ListPayload) (res *gen.AssignmentList, err error) {
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

	if p.EquipmentID != nil {
		query += ` AND a.equipment_id = $` + strconv.Itoa(argIdx)
		args = append(args, *p.EquipmentID)
		argIdx++
	}
	if p.IsActive != nil {
		query += ` AND a.is_active = $` + strconv.Itoa(argIdx)
		args = append(args, *p.IsActive)
		argIdx++
	}

	query += ` ORDER BY a.assigned_at DESC`

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []*gen.Assignment
	for rows.Next() {
		a := &gen.Assignment{}
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
	return &gen.AssignmentList{Assignments: assignments}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Assignment, err error) {
	res = &gen.Assignment{}
	err = s.db.QueryRow(ctx, `SELECT
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
	WHERE a.id = $1`, p.ID).Scan(
		&res.TargetType,
		&res.CardNumber,
		&res.FullName,
		&res.WaybillNumber,
		&res.WaybillDate,
		&res.FromDeptName,
		&res.ToDeptName,
		&res.AssignedAt,
		&res.UnassignedAt,
		&res.OperatorComment,
	)
	return
}
