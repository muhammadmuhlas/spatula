package listing

// Repository provides access to the beer and review storage.
type Repository interface {

	// GetIssue returns the issue with given issue keys.
	GetIssue(string) (Issue, error)

	// GetAllIssues returns all issues using jql.
	GetAllIssues(string, int, bool) ([]Issue, error)
}

// Service provides beer and review listing operations.
type Service interface {
	GetIssue(string) (Issue, error)
	GetIssues(string, int, bool) ([]Issue, error)
}

type service struct {
	r Repository
}

// NewService creates a listing service with the necessary dependencies
func NewService(r Repository) Service {
	return &service{r}
}

// GetIssues returns all issues
func (s *service) GetIssues(jql string, size int, dev bool) ([]Issue, error) {
	return s.r.GetAllIssues(jql, size, dev)
}

// GetIssue returns an issue
func (s *service) GetIssue(id string) (Issue, error) {
	return s.r.GetIssue(id)
}
