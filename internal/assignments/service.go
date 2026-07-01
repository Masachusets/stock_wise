package assignments

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/assignments"
)

type service struct {
	repo Repository
}

func New(repo Repository) gen.Service {
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, p *gen.ListPayload) (res *gen.AssignmentList, err error) {
	items, err := s.repo.List(ctx, p.EquipmentID, p.IsActive)
	if err != nil {
		return nil, err
	}

	var assignments []*gen.Assignment
	for _, a := range items {
		assignments = append(assignments, &gen.Assignment{
			TargetType:      a.TargetType,
			CardNumber:      a.CardNumber,
			FullName:        a.FullName,
			DeptName:        a.DeptName,
			AssignedAt:      a.AssignedAt,
			UnassignedAt:    a.UnassignedAt,
			OperatorComment: a.OperatorComment,
		})
	}
	return &gen.AssignmentList{Assignments: assignments}, nil
}

func (s *service) Get(ctx context.Context, p *gen.GetPayload) (res *gen.Assignment, err error) {
	a, err := s.repo.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return &gen.Assignment{
		TargetType:      a.TargetType,
		CardNumber:      a.CardNumber,
		FullName:        a.FullName,
		DeptName:        a.DeptName,
		AssignedAt:      a.AssignedAt,
		UnassignedAt:    a.UnassignedAt,
		OperatorComment: a.OperatorComment,
	}, nil
}
