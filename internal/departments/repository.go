package departments

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// postgresRepository реализует Repository для PostgreSQL.
type postgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository создаёт новый репозиторий.
func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) List(ctx context.Context) ([]*Department, error) {
	rows, err := r.db.Query(ctx, "SELECT code, type, name FROM departments ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []*Department
	for rows.Next() {
		d := &Department{}
		if err := rows.Scan(&d.Code, &d.Type, &d._Name); err != nil {
			return nil, err
		}
		departments = append(departments, d)
	}
	return departments, nil
}

func (r *postgresRepository) Get(ctx context.Context, code int32) (*Department, error) {
	d := &Department{}
	err := r.db.QueryRow(ctx, "SELECT code, type, name FROM departments WHERE code = $1", code).Scan(&d.Code, &d.Type, &d._Name)
	return d, err
}
