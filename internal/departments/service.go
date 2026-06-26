package departments

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/departments"
)

type service struct {
	repo Repository
}

// New создаёт сервис departments.
func New(repo Repository) gen.Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context) (res *gen.DepartmentList, err error) {
	depts, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var departments []*gen.Department
	for _, d := range depts {
		departments = append(departments, &gen.Department{
			Code: d.Code,
			Type: d.Type,
			Name: d.Name(),
		})
	}
	return &gen.DepartmentList{Departments: departments}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Department, err error) {
	d, err := s.repo.Get(ctx, p.Code)
	if err != nil {
		return nil, err
	}
	return &gen.Department{
		Code: d.Code,
		Type: d.Type,
		Name: d.Name(),
	}, nil
}
