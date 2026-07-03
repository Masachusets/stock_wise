package waybills

import (
	"context"

	gen "github.com/Masachusets/stock_wise/gen/waybills"
)

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
	CreateAssignment(ctx context.Context, equipmentID int32, departmentCode int32) error
	Delete(ctx context.Context, id int32) error
	GetStatus(ctx context.Context, id int32) (string, error)
	ListForWeb(ctx context.Context) ([]*WaybillListItem, error)
	GetForWeb(ctx context.Context, id int32) (*WaybillDetail, error)
}

// WaybillListItem — элемент списка накладных для web-интерфейса.
type WaybillListItem struct {
	ID        int32
	Number    string
	IssueDate string
	Status    string
	FromName  *string
	ToName    *string
}

// WaybillDetail — детальная информация о накладной для web-интерфейса.
type WaybillDetail struct {
	Number     string
	IssueDate  string
	Status     string
	FromName   *string
	ToName     *string
	Equipments []WaybillEquipmentItem
}

// WaybillEquipmentItem — оборудование в составе накладной.
type WaybillEquipmentItem struct {
	InventoryNumber string
	ModelName       string
	Nomenclature    string
}

// WebService интерфейс для web-хэндлеров.
type WebService interface {
	ListForWeb(ctx context.Context) ([]*WaybillListItem, error)
	GetForWeb(ctx context.Context, id int32) (*WaybillDetail, error)
	Sign(ctx context.Context, p *gen.SignPayload) (*gen.Waybill, error)
	Archive(ctx context.Context, p *gen.ArchivePayload) (*gen.Waybill, error)
	Delete(ctx context.Context, p *gen.DeletePayload) error
}

// SignPayload payload for signing a waybill.
type SignPayload struct {
	ID int32
}

// ArchivePayload payload for archiving a waybill.
type ArchivePayload struct {
	ID int32
}

// DeletePayload payload for deleting a waybill.
type DeletePayload struct {
	ID int32
}
