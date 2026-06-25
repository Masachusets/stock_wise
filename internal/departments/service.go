package departments

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/departments"
	"github.com/jackc/pgx/v5/pgxpool"
)

type service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) gen.Service {
	return &service{db: db}
}

func (s *service) List(ctx context.Context) (res *gen.DepartmentList, err error) {
	rows, err := s.db.Query(ctx, "SELECT code, type, name FROM departments ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*gen.Department
	for rows.Next() {
		d := &gen.Department{}
		if err := rows.Scan(&d.Code, &d.Type, &d.Name); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return &gen.DepartmentList{Departments: departments}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Department, err error) {
	res = &gen.Department{}
	err = s.db.QueryRow(ctx, "SELECT code, type, name FROM departments WHERE code = $1", p.Code).Scan(&res.Code, &res.Type, &res.Name)
	return
}
