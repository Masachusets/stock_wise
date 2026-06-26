package nomenclatures

import "context"

// Nomenclature — сущность номенклатуры.
type Nomenclature struct {
	ID   int32
	Code string
	Name string
}

// Repository интерфейс для доступа к данным номенклатур.
type Repository interface {
	List(ctx context.Context) ([]*Nomenclature, error)
	Get(ctx context.Context, id int32) (*Nomenclature, error)
}
