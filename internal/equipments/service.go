package equipments

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/equipments"
)

type service struct {
	repo Repository
}

func New(repo Repository) gen.Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, p *gen.ListPayload) (res *gen.EquipmentList, err error) {
	filter := &ListFilter{
		Status:         p.Status,
		NomenclatureID: p.NomenclatureID,
		Location:       p.Location,
		Search:         p.Search,
	}

	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var eqs []*gen.Equipment
	for _, e := range items {
		eqs = append(eqs, convertToGen(e))
	}
	return &gen.EquipmentList{Equipments: eqs}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Equipment, err error) {
	e, err := s.repo.Get(ctx, p.InventoryNumber)
	if err != nil {
		return nil, err
	}
	return convertToGen(e), nil
}

func (s *service) Create(ctx context.Context, p *gen.CreateEquipmentPayload) (res *gen.Equipment, err error) {
	e := &Equipment{
		InventoryNumber: p.InventoryNumber,
		SerialNumber:    p.SerialNumber,
		ModelName:       p.ModelName,
		ManufactureDate: p.ManufactureDate,
		ArrivalDate:     p.ArrivalDate,
		Status:          p.Status,
		FormNumber:      p.FormNumber,
		Location:        p.Location,
		Notes:           p.Notes,
	}
	if err := s.repo.Create(ctx, e); err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{InventoryNumber: p.InventoryNumber})
}

func (s *service) Update(ctx context.Context, p *gen.UpdateEquipmentPayload) (res *gen.Equipment, err error) {
	e, err := s.repo.Get(ctx, p.InventoryNumber)
	if err != nil {
		return nil, err
	}

	if p.SerialNumber != nil {
		e.SerialNumber = p.SerialNumber
	}
	if p.ModelName != nil {
		e.ModelName = *p.ModelName
	}
	if p.ManufactureDate != nil {
		e.ManufactureDate = p.ManufactureDate
	}
	if p.ArrivalDate != nil {
		e.ArrivalDate = p.ArrivalDate
	}
	if p.Status != nil {
		e.Status = *p.Status
	}
	if p.FormNumber != nil {
		e.FormNumber = p.FormNumber
	}
	if p.Location != nil {
		e.Location = p.Location
	}
	if p.Notes != nil {
		e.Notes = p.Notes
	}

	if err := s.repo.Update(ctx, e); err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{InventoryNumber: p.InventoryNumber})
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	return s.repo.Delete(ctx, p.InventoryNumber)
}

func convertToGen(e *Equipment) *gen.Equipment {
	g := &gen.Equipment{
		InventoryNumber: e.InventoryNumber,
		ModelName:       e.ModelName,
		Status:          e.Status,
		SerialNumber:    e.SerialNumber,
		ManufactureDate: e.ManufactureDate,
		ArrivalDate:     e.ArrivalDate,
		FormNumber:      e.FormNumber,
		Location:        e.Location,
		Notes:           e.Notes,
	}
	if e.Nomenclature != nil {
		g.Nomenclature = &gen.Nomenclature{
			Code: e.Nomenclature.Code,
			Name: e.Nomenclature.Name,
		}
	}
	if e.Assignment != nil {
		g.Assignment = &gen.AssignmentInfo{
			TargetType:      e.Assignment.TargetType,
			CardNumber:      e.Assignment.CardNumber,
			FullName:        e.Assignment.FullName,
			WaybillNumber:   e.Assignment.WaybillNumber,
			WaybillDate:     e.Assignment.WaybillDate,
			FromDeptName:    e.Assignment.FromDeptName,
			ToDeptName:      e.Assignment.ToDeptName,
			OperatorComment: e.Assignment.OperatorComment,
		}
	}
	return g
}
