package cards

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/cards"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context) (res *gen.CardList, err error) {
	rows, err := s.db.Query(ctx, "SELECT number, full_name FROM cards ORDER BY number")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*gen.Card
	for rows.Next() {
		c := &gen.Card{}
		if err := rows.Scan(&c.Number, &c.FullName); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return &gen.CardList{Cards: cards}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Card, err error) {
	res = &gen.Card{}
	err = s.db.QueryRow(ctx, "SELECT number, full_name FROM cards WHERE number = $1", p.Number).Scan(&res.Number, &res.FullName)
	return
}

func (s *service) Create(ctx context.Context, p *gen.CreateCardPayload) (res *gen.Card, err error) {
	_, err = s.db.Exec(ctx, "INSERT INTO cards (number, full_name) VALUES ($1, $2)", p.Number, p.FullName)
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{Number: p.Number})
}

func (s *service) Update(ctx context.Context, p *gen.UpdateCardPayload) (res *gen.Card, err error) {
	if p.FullName != nil {
		_, err = s.db.Exec(ctx, "UPDATE cards SET full_name = $1 WHERE number = $2", *p.FullName, p.Number)
	} else {
		_, err = s.db.Exec(ctx, "UPDATE cards SET updated_at = NOW() WHERE number = $1", p.Number)
	}
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, &gen.GetPayload{Number: p.Number})
}

func (s *service) Delete(ctx context.Context, p *gen.DeletePayload) error {
	_, err := s.db.Exec(ctx, "DELETE FROM cards WHERE number = $1", p.Number)
	return err
}
