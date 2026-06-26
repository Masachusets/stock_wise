package cards

import "context"

// Card — сущность карточки сотрудника.
type Card struct {
	Number   int32
	FullName string
}

// Repository интерфейс для доступа к данным карточек.
type Repository interface {
	List(ctx context.Context) ([]*Card, error)
	Get(ctx context.Context, number int32) (*Card, error)
	Create(ctx context.Context, card *Card) error
	Update(ctx context.Context, card *Card) error
	Delete(ctx context.Context, number int32) error
}
