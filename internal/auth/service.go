package auth

import (
	"context"

	"github.com/peterbourgon/sympatico/internal/ctxlog"
)

// Service provides the API.
type Service struct {
	repo Repository
}

// Repository is a client-side interface, which models
// the concrete e.g. SQLiteRepository.
type Repository interface {
	Create(ctx context.Context, user, pass string) error
	Auth(ctx context.Context, user, pass string) (token string, err error)
	Deauth(ctx context.Context, user, token string) error
	Validate(ctx context.Context, user, token string) error
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
func (s *Service) Signup(ctx context.Context, user, pass string) (err error) {
	defer func() { ctxlog.From(ctx).Log("auth_method", "Signup", "signup_user", user, "signup_err", err) }()
	return s.repo.Create(ctx, user, pass)
}

// Login logs the user in, if the pass is correct.
// The returned token should be passed to Logout or Validate.
func (s *Service) Login(ctx context.Context, user, pass string) (token string, err error) {
	defer func() { ctxlog.From(ctx).Log("auth_method", "Login", "login_user", user, "login_err", err) }()
	return s.repo.Auth(ctx, user, pass)
}

// Logout logs the user out, if the token is valid.
func (s *Service) Logout(ctx context.Context, user, token string) (err error) {
	defer func() { ctxlog.From(ctx).Log("auth_method", "Logout", "logout_user", user, "logout_err", err) }()
	return s.repo.Deauth(ctx, user, token)
}

// Validate returns a nil error if the user is logged in and
// provides the correct token.
func (s *Service) Validate(ctx context.Context, user, token string) (err error) {
	defer func() { ctxlog.From(ctx).Log("auth_method", "Validate", "validate_user", user, "validate_err", err) }()
	return s.repo.Validate(ctx, user, token)
}
