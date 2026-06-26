package departments

import "context"

// Department — сущность подразделения.
type Department struct {
	Code int32
	Type string
	_Name string
}

// Name возвращает наименование подразделения.
func (d *Department) Name() string { return d._Name }

// SetName устанавливает наименование подразделения.
func (d *Department) SetName(v string) { d._Name = v }

// Repository интерфейс для доступа к данным подразделений.
type Repository interface {
	List(ctx context.Context) ([]*Department, error)
	Get(ctx context.Context, code int32) (*Department, error)
}
