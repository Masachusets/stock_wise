package cards

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context) ([]*Card, error) {
	rows, err := r.db.Query(ctx, "SELECT number, full_name FROM cards ORDER BY number")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*Card
	for rows.Next() {
		c := &Card{}
		if err := rows.Scan(&c.Number, &c.FullName); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *postgresRepository) Get(ctx context.Context, number int32) (*Card, error) {
	c := &Card{}
	err := r.db.QueryRow(ctx, "SELECT number, full_name FROM cards WHERE number = $1", number).Scan(&c.Number, &c.FullName)
	return c, err
}

func (r *postgresRepository) Create(ctx context.Context, card *Card) error {
	_, err := r.db.Exec(ctx, "INSERT INTO cards (number, full_name) VALUES ($1, $2)", card.Number, card.FullName)
	return err
}

func (r *postgresRepository) Update(ctx context.Context, card *Card) error {
	_, err := r.db.Exec(ctx, "UPDATE cards SET full_name = $1, updated_at = NOW() WHERE number = $2", card.FullName, card.Number)
	return err
}

func (r *postgresRepository) Delete(ctx context.Context, number int32) error {
	_, err := r.db.Exec(ctx, "DELETE FROM cards WHERE number = $1", number)
	return err
}
