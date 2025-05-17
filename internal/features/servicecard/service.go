package servicecard

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTicket(t *Ticket) error {
	return s.repo.Create(t)
}

func (s *Service) GetTicket(id string) (*Ticket, error) {
	return s.repo.GetByID(id)
}

func (s *Service) ListUserTickets(userID string) ([]Ticket, error) {
	return s.repo.ListByUser(userID)
}

func (s *Service) UpdateTicket(t *Ticket) error {
	return s.repo.Update(t)
}
