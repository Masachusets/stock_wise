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
}
