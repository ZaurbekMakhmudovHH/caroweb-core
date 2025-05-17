package servicecard

type Repository interface {
	Create(ticket *Ticket) error
	GetByID(id string) (*Ticket, error)
	ListByUser(userID string) ([]Ticket, error)
	Update(ticket *Ticket) error
}
