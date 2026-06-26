package waybills

import "context"

// Waybill — сущность накладной.
type Waybill struct {
	ID        int32
	Number    string
	IssueDate string
	FromDept  *int32
	ToDept    *int32
	Status    string
	Items     []*WaybillsEquipment
}

// WaybillsEquipment — позиция накладной.
type WaybillsEquipment struct {
	WaybillID   int32
	EquipmentID int32
}

// Repository интерфейс для доступа к данным накладных.
type Repository interface {
	List(ctx context.Context) ([]*Waybill, error)
	Get(ctx context.Context, id int32) (*Waybill, error)
	Create(ctx context.Context, wb *Waybill) error
	GetEquipmentIDs(ctx context.Context, waybillID int32) ([]int32, error)
	UpdateStatus(ctx context.Context, id int32, status string) error
	CreateAssignment(ctx context.Context, equipmentID int32, waybillID int32) error
	Delete(ctx context.Context, id int32) error
	GetStatus(ctx context.Context, id int32) (string, error)
}
