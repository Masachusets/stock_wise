package waybills

import (
	"context"
	"fmt"

	gen "github.com/Masachusets/stock_wise/gen/waybills"
)

type service struct {
	repo Repository
}

func New(repo Repository) *service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) (res *gen.WaybillList, err error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var waybills []*gen.Waybill
	for _, w := range items {
		waybills = append(waybills, &gen.Waybill{
			Number:    w.Number,
			IssueDate: w.IssueDate,
			FromDept:  w.FromDept,
			ToDept:    w.ToDept,
			Status:    w.Status,
		})
	}
	return &gen.WaybillList{Waybills: waybills}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Waybill, err error) {
	w, err := s.repo.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	res = &gen.Waybill{
		Number:    w.Number,
		IssueDate: w.IssueDate,
		FromDept:  w.FromDept,
		ToDept:    w.ToDept,
		Status:    w.Status,
	}

	for _, item := range w.Items {
		res.Items = append(res.Items, &gen.WaybillsEquipment{
			WaybillID:   item.WaybillID,
			EquipmentID: item.EquipmentID,
		})
	}
	return res, nil
}

func (s *service) Create(ctx context.Context, p *gen.CreateWaybillPayload) (res *gen.Waybill, err error) {
	wb := &Waybill{
		Number:    p.Number,
		IssueDate: p.IssueDate,
		FromDept:  p.FromDept,
		ToDept:    p.ToDept,
	}
	for _, item := range p.Items {
		wb.Items = append(wb.Items, &WaybillsEquipment{
			EquipmentID: item.EquipmentID,
		})
	}

	if err := s.repo.Create(ctx, wb); err != nil {
		return nil, err
	}

	return s.Get(ctx, &gen.GetPayload{ID: wb.ID})
}

func (s *service) Sign(ctx context.Context, p *gen.SignPayload) (res *gen.Waybill, err error) {
	status, err := s.repo.GetStatus(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if status != "draft" {
		return nil, gen.InvalidStatus(fmt.Sprintf("нельзя подписать накладную со статусом %s", status))
	}

	if err := s.repo.UpdateStatus(ctx, p.ID, "signed"); err != nil {
		return nil, err
	}

	// Получить to_dept из накладной
	wb, err := s.repo.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}

	eqIDs, err := s.repo.GetEquipmentIDs(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	for _, eqID := range eqIDs {
		if wb.ToDept != nil {
			if err := s.repo.CreateAssignment(ctx, eqID, *wb.ToDept); err != nil {
				return nil, err
			}
		}
	}

	return s.Get(ctx, &gen.GetPayload{ID: p.ID})
}

func (s *service) Archive(ctx context.Context, p *gen.ArchivePayload) (res *gen.Waybill, err error) {
	status, err := s.repo.GetStatus(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if status != "signed" {
		return nil, gen.InvalidStatus(fmt.Sprintf("нельзя архивировать накладную со статусом %s", status))
	}

	if err := s.repo.UpdateStatus(ctx, p.ID, "archived"); err != nil {
		return nil, err
	}

	return s.Get(ctx, &gen.GetPayload{ID: p.ID})
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	status, err := s.repo.GetStatus(ctx, p.ID)
	if err != nil {
		return err
	}
	if status != "draft" {
		return gen.InvalidStatus(fmt.Sprintf("нельзя удалить накладную со статусом %s", status))
	}

	return s.repo.Delete(ctx, p.ID)
}

func (s *service) ListForWeb(ctx context.Context) ([]*WaybillListItem, error) {
	return s.repo.ListForWeb(ctx)
}

func (s *service) GetForWeb(ctx context.Context, id int32) (*WaybillDetail, error) {
	return s.repo.GetForWeb(ctx, id)
}
