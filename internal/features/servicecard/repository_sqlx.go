package servicecard

import "github.com/jmoiron/sqlx"

type SQLXRepository struct {
	db *sqlx.DB
}

func NewSQLXRepository(db *sqlx.DB) *SQLXRepository {
	return &SQLXRepository{db: db}
}

func (r *SQLXRepository) Create(ticket *Ticket) error {
	query := `INSERT INTO tickets (id, title, content, status, user_id, created_at) 
	VALUES (:id, :title, :content, :status, :user_id, :created_at)`
	_, err := r.db.NamedExec(query, ticket)
	return err
}

func (r *SQLXRepository) GetByID(id string) (*Ticket, error) {
	var ticket Ticket
	err := r.db.Get(&ticket, "SELECT * FROM tickets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *SQLXRepository) ListByUser(userID string) ([]Ticket, error) {
	var tickets []Ticket
	err := r.db.Select(&tickets, "SELECT * FROM tickets WHERE user_id = $1", userID)
	return tickets, err
}

func (r *SQLXRepository) Update(ticket *Ticket) error {
	query := `UPDATE tickets SET title = :title, content = :content, status = :status WHERE id = :id`
	_, err := r.db.NamedExec(query, ticket)
	return err
}
