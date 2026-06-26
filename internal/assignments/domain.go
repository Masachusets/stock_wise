package assignments

import "context"

// Assignment — сущность закрепления оборудования.
type Assignment struct {
	TargetType      string
	CardNumber      *int32
	FullName        *string
	WaybillNumber   *string
	WaybillDate     *string
	FromDeptName    *string
	ToDeptName      *string
	AssignedAt      string
	UnassignedAt    *string
	OperatorComment *string
}

// Repository интерфейс для доступа к данным закреплений.
type Repository interface {
	List(ctx context.Context, equipmentID *int32, isActive *bool) ([]*Assignment, error)
	Get(ctx context.Context, id int32) (*Assignment, error)
}
