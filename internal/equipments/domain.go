package equipments

import "context"

// Equipment — сущность оборудования.
type Equipment struct {
	InventoryNumber string
	SerialNumber    *string
	NomenclatureID  *int32
	Nomenclature    *NomenclatureInfo
	ModelName       string
	ManufactureDate *string
	ArrivalDate     *string
	Status          string
	FormNumber      *string
	Location        *string
	Notes           *string
	Assignment      *AssignmentInfo
}

// NomenclatureInfo — информация о номенклатуре.
type NomenclatureInfo struct {
	Code string
	Name string
}

// AssignmentInfo — информация о закреплении.
type AssignmentInfo struct {
	TargetType      string
	CardNumber      *int32
	FullName        *string
	DeptName        *string
	OperatorComment *string
}

// ListFilter — фильтр для списка оборудования.
type ListFilter struct {
	Status          *string
	NomenclatureID  *int32
	Location        *string
	Search          *string
}

// Repository интерфейс для доступа к данным оборудования.
type Repository interface {
	List(ctx context.Context, filter *ListFilter) ([]*Equipment, error)
	Get(ctx context.Context, inventoryNumber string) (*Equipment, error)
	Create(ctx context.Context, eq *Equipment) error
	Update(ctx context.Context, eq *Equipment) error
	Delete(ctx context.Context, inventoryNumber string) error
	ListForWeb(ctx context.Context, filter *ListFilter) ([]*EquipmentListItem, error)
	GetForWeb(ctx context.Context, inventoryNumber string) (*EquipmentDetail, error)
	ListNomenclatures(ctx context.Context) ([]*NomenclatureOption, error)
	CreateWithAssignment(ctx context.Context, eq *Equipment, departmentCode int) error
	ListDeleted(ctx context.Context) ([]*EquipmentDeletedItem, error)
}

// EquipmentListItem — элемент списка оборудования для web-интерфейса.
type EquipmentListItem struct {
	InventoryNumber string
	ModelName       string
	Status          string
	Location        *string
	Nomenclature    *NomenclatureInfo
	Assignment      *AssignmentInfo
}

// EquipmentDetail — детальная информация об оборудовании для web-интерфейса.
type EquipmentDetail struct {
	InventoryNumber string
	SerialNumber    *string
	ModelName       string
	ManufactureDate *string
	ArrivalDate     *string
	Status          string
	FormNumber      *string
	Location        *string
	Notes           *string
	Nomenclature    *NomenclatureInfo
	Assignment      *AssignmentInfo
}

// EquipmentDeletedItem — удалённое оборудование.
type EquipmentDeletedItem struct {
	InventoryNumber string
	ModelName       string
	NomCode         string
	NomName         string
	Status          string
	DeletedAt       string
}

// NomenclatureOption — опция для выпадающего списка номенклатур.
type NomenclatureOption struct {
	ID   int32
	Code string
	Name string
}

// WebService интерфейс для web-хэндлеров оборудования.
type WebService interface {
	ListForWeb(ctx context.Context, filter *ListFilter) ([]*EquipmentListItem, error)
	GetForWeb(ctx context.Context, inventoryNumber string) (*EquipmentDetail, error)
	ListNomenclatures(ctx context.Context) ([]*NomenclatureOption, error)
	CreateWithAssignment(ctx context.Context, eq *Equipment, departmentCode int) error
	Update(ctx context.Context, eq *Equipment) error
	Delete(ctx context.Context, inventoryNumber string) error
	ListDeleted(ctx context.Context) ([]*EquipmentDeletedItem, error)
}
