package pathtask

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	Create(from, to string) (*Task, error)
	Get(id uuid.UUID) (*Task, error)
}

type Repo struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) Create(from, to string) (*Task, error) {
	task := &Task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FromURL:   from,
		ToURL:     to,
		Status:    StatusPending,
	}

	query := `INSERT INTO "tasks" (id, created_at, updated_at, from_url, to_url, status) 
							VALUES (:id, :created_at, :updated_at, :from_url, :to_url, :status)`
	_, err := r.db.NamedExec(query, task)
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}

func (r *Repo) Get(id uuid.UUID) (*Task, error) {
	task := new(Task)
	err := r.db.Get(task, `SELECT * FROM "tasks" WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "database error")
	}

	return task, nil
}
