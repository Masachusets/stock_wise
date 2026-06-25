package nomenclatures

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/nomenclatures"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context) (res *gen.NomenclatureList, err error) {
	rows, err := s.db.Query(ctx, "SELECT code, name FROM nomenclatures ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nomenclatures []*gen.Nomenclature
	for rows.Next() {
		n := &gen.Nomenclature{}
		if err := rows.Scan(&n.Code, &n.Name); err != nil {
			return nil, err
		}
		nomenclatures = append(nomenclatures, n)
	}
	return &gen.NomenclatureList{Nomenclatures: nomenclatures}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Nomenclature, err error) {
	res = &gen.Nomenclature{}
	err = s.db.QueryRow(ctx, "SELECT code, name FROM nomenclatures WHERE id = $1", p.ID).Scan(&res.Code, &res.Name)
	return
}
