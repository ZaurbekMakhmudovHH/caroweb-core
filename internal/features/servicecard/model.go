package servicecard

type Ticket struct {
	ID        string `db:"id" json:"id"`
	Title     string `db:"title" json:"title"`
	Content   string `db:"content" json:"content"`
	Status    string `db:"status" json:"status"`
	UserID    string `db:"user_id" json:"user_id"`
	CreatedAt string `db:"created_at" json:"created_at"`
}

const (
	StatusOpen   = "open"
	StatusClosed = "closed"
)
