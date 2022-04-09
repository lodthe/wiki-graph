package pathtask

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status uint

const (
	StatusPending Status = iota + 1
	StatusProcessing
	StatusDone
)

type Task struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	FromURL string `db:"from_url"`
	ToURL   string `db:"to_url"`

	Status Status `db:"status"`

	Result *Result `db:"result"`
}

type Result struct {
	ShortestPath []string `json:"shortest_path"`
}

func (r *Result) Value() (driver.Value, error) {
	return json.Marshal(*r)
}

func (r *Result) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("value cannot be converted to []byte")
	}

	return json.Unmarshal(b, r)
}
