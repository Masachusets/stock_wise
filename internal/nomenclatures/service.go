package nomenclatures

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/nomenclatures"
)

type service struct {
	repo Repository
}

func New(repo Repository) gen.Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) (res *gen.NomenclatureList, err error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var nomenclatures []*gen.Nomenclature
	for _, n := range items {
		nomenclatures = append(nomenclatures, &gen.Nomenclature{
			Code: n.Code,
			Name: n.Name,
		})
	}
	return &gen.NomenclatureList{Nomenclatures: nomenclatures}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Nomenclature, err error) {
	n, err := s.repo.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return &gen.Nomenclature{
		Code: n.Code,
		Name: n.Name,
	}, nil
}
