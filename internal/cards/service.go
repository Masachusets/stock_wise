package cards

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/cards"
)

type service struct {
	repo Repository
}

func New(repo Repository) gen.Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) (res *gen.CardList, err error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var cards []*gen.Card
	for _, c := range items {
		cards = append(cards, &gen.Card{
			Number:   c.Number,
			FullName: c.FullName,
		})
	}
	return &gen.CardList{Cards: cards}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Card, err error) {
	c, err := s.repo.Get(ctx, p.Number)
	if err != nil {
		return nil, err
	}
	return &gen.Card{
		Number:   c.Number,
		FullName: c.FullName,
	}, nil
}

func (s *service) Create(ctx context.Context, p *gen.CreateCardPayload) (res *gen.Card, err error) {
	c := &Card{Number: p.Number, FullName: p.FullName}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return &gen.Card{Number: c.Number, FullName: c.FullName}, nil
}

func (s *service) Update(ctx context.Context, p *gen.UpdateCardPayload) (res *gen.Card, err error) {
	c, err := s.repo.Get(ctx, p.Number)
	if err != nil {
		return nil, err
	}
	if p.FullName != nil {
		c.FullName = *p.FullName
	}
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return &gen.Card{Number: c.Number, FullName: c.FullName}, nil
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	return s.repo.Delete(ctx, p.Number)
}
