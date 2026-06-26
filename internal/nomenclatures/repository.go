package nomenclatures

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

func (r *postgresRepository) List(ctx context.Context) ([]*Nomenclature, error) {
	rows, err := r.db.Query(ctx, "SELECT id, code, name FROM nomenclatures ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nomenclatures []*Nomenclature
	for rows.Next() {
		n := &Nomenclature{}
		if err := rows.Scan(&n.ID, &n.Code, &n.Name); err != nil {
			return nil, err
		}
		nomenclatures = append(nomenclatures, n)
	}
	return nomenclatures, nil
}

func (r *postgresRepository) Get(ctx context.Context, id int32) (*Nomenclature, error) {
	n := &Nomenclature{}
	err := r.db.QueryRow(ctx, "SELECT id, code, name FROM nomenclatures WHERE id = $1", id).Scan(&n.ID, &n.Code, &n.Name)
	return n, err
}
