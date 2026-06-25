package waybills

import (
	"context"
	"fmt"

	gen "github.com/Masachusets/stock_wise/gen/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context) (res *gen.WaybillList, err error) {
	rows, err := s.db.Query(ctx, "SELECT id, number, issue_date::text, from_dept, to_dept, status FROM waybills WHERE deleted_at IS NULL ORDER BY issue_date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var waybills []*gen.Waybill
	for rows.Next() {
		w := &gen.Waybill{}
		var id int32
		if err := rows.Scan(&id, &w.Number, &w.IssueDate, &w.FromDept, &w.ToDept, &w.Status); err != nil {
			return nil, err
		}
		waybills = append(waybills, w)
	}
	return &gen.WaybillList{Waybills: waybills}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Waybill, err error) {
	res = &gen.Waybill{}
	err = s.db.QueryRow(ctx, "SELECT number, issue_date::text, from_dept, to_dept, status FROM waybills WHERE id = $1", p.ID).Scan(&res.Number, &res.IssueDate, &res.FromDept, &res.ToDept, &res.Status)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, "SELECT waybill_id, equipment_id FROM waybills_equipments WHERE waybill_id = $1", p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		we := &gen.WaybillsEquipment{}
		if err := rows.Scan(&we.WaybillID, &we.EquipmentID); err != nil {
			return nil, err
		}
		res.Items = append(res.Items, we)
	}
	return
}

func (s *service) Create(ctx context.Context, p *gen.CreateWaybillPayload) (res *gen.Waybill, err error) {
	var id int32
	err = s.db.QueryRow(ctx,
		"INSERT INTO waybills (number, issue_date, from_dept, to_dept) VALUES ($1, $2, $3, $4) RETURNING id",
		p.Number, p.IssueDate, p.FromDept, p.ToDept,
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	for _, item := range p.Items {
		_, err = s.db.Exec(ctx, "INSERT INTO waybills_equipments (waybill_id, equipment_id) VALUES ($1, $2)", id, item.EquipmentID)
		if err != nil {
			return nil, err
		}
	}

	return s.Get(ctx, &gen.GetPayload{ID: id})
}

func (s *service) Sign(ctx context.Context, p *gen.SignPayload) (res *gen.Waybill, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM waybills WHERE id = $1 FOR UPDATE", p.ID).Scan(&status)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, gen.InvalidStatus(fmt.Sprintf("нельзя подписать накладную со статусом %s", status))
	}

	_, err = tx.Exec(ctx, "UPDATE waybills SET status = 'signed' WHERE id = $1", p.ID)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, "SELECT equipment_id FROM waybills_equipments WHERE waybill_id = $1", p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var eqID int32
		if err := rows.Scan(&eqID); err != nil {
			return nil, err
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO equipments_assignments (equipment_id, target_type, waybill_id)
			 VALUES ($1, 'department', $2)`, eqID, p.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.Get(ctx, &gen.GetPayload{ID: p.ID})
}

func (s *service) Archive(ctx context.Context, p *gen.ArchivePayload) (res *gen.Waybill, err error) {
	var status string
	err = s.db.QueryRow(ctx, "SELECT status FROM waybills WHERE id = $1", p.ID).Scan(&status)
	if err != nil {
		return nil, err
	}
	if status != "signed" {
		return nil, gen.InvalidStatus(fmt.Sprintf("нельзя архивировать накладную со статусом %s", status))
	}

	_, err = s.db.Exec(ctx, "UPDATE waybills SET status = 'archived' WHERE id = $1", p.ID)
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, &gen.GetPayload{ID: p.ID})
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	var status string
	err := s.db.QueryRow(ctx, "SELECT status FROM waybills WHERE id = $1", p.ID).Scan(&status)
	if err != nil {
		return err
	}
	if status != "draft" {
		return gen.InvalidStatus(fmt.Sprintf("нельзя удалить накладную со статусом %s", status))
	}

	_, err = s.db.Exec(ctx, "DELETE FROM waybills WHERE id = $1", p.ID)
	return err
}
