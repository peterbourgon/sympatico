package auth

// Service provides the API.
type Service struct {
	repo Repository
}

// Repository is a client-side interface, which models
// the concrete e.g. SQLiteRepository.
type Repository interface {
	Create(user, pass string) error
	Auth(user, pass string) (token string, err error)
	Deauth(user, token string) error
	Validate(user, token string) error
}

// NewService returns a usable service, wrapping a repository.
// Most auth logic occurs in the repository layer,
// so this is a very thin wrapper.
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Signup creates a user with the given pass.
// The user still needs to login.
func (s *Service) Signup(user, pass string) error {
	return s.repo.Create(user, pass)
}

// Login logs the user in, if the pass is correct.
// The returned token should be passed to Logout or Validate.
func (s *Service) Login(user, pass string) (token string, err error) {
	return s.repo.Auth(user, pass)
}

// Logout logs the user out, if the token is valid.
func (s *Service) Logout(user, token string) error {
	return s.repo.Deauth(user, token)
}

// Validate returns a nil error if the user is logged in and
// provides the correct token.
func (s *Service) Validate(user, token string) error {
	return s.repo.Validate(user, token)
}
