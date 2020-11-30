package working

// Repository provides access to the beer and review storage.
type Repository interface {

	// GetIssue returns the issue with given issue keys.
	CreatePullRequest(request PullRequestTask) (interface{}, error)

	CreateBranch(owner, repository, name, source string) (Branch, error)

	GetBranch(owner, repository, name string) (Branch, error)
}

// Service provides beer and review listing operations.
type Service interface {
	CreatePullRequest(request PullRequestTask) (interface{}, error)
	CreateBranch(owner, repository, name, source string) (Branch, error)
	GetBranch(owner, repository, name string) (Branch, error)
	BranchExist(owner, repository, name string) bool
}

type service struct {
	r Repository
}

// NewService creates a listing service with the necessary dependencies
func NewService(r Repository) Service {
	return &service{r}
}

// GetIssues returns all issues
func (s *service) CreatePullRequest(opt PullRequestTask) (interface{}, error) {
	return s.r.CreatePullRequest(opt)
}

func (s *service) CreateBranch(owner, repository, name, source string) (Branch, error) {
	return s.r.CreateBranch(owner, repository, name, source)
}

func (s *service) GetBranch(owner, repository, name string) (Branch, error) {
	return s.r.GetBranch(owner, repository, name)
}

func (s *service) BranchExist(owner, repository, name string) bool {
	if _, err := s.r.GetBranch(owner, repository, name); err != nil {
		return false
	}
	return true
}
