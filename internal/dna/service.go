package dna

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/pkg/errors"

	"github.com/peterbourgon/sympatico/internal/ctxlog"
)

var (
	// ErrSubsequenceNotFound is returned by Check on a failure.
	ErrSubsequenceNotFound = errors.New("subsequence doesn't appear in the DNA sequence")

	// ErrBadAuth is returned if a user validation check fails.
	ErrBadAuth = errors.New("bad auth")

	// ErrInvalidSequence is returned if an invalid sequence is added.
	ErrInvalidSequence = errors.New("invalid DNA sequence")
)

// Service provides the API.
type Service struct {
	repo   Repository
	valid  Validator
	logger log.Logger
}

// Repository is a client-side interface, which models
// the concrete e.g. SQLiteRepository.
//
// We use an interface instead of the concrete type,
// so we can more easily test the Service.
type Repository interface {
	Insert(ctx context.Context, user, sequence string) error
	Select(ctx context.Context, user string) (sequence string, err error)
}

// Validator is a client-side interface, which models
// the parts of the auth service that we use.
type Validator interface {
	Validate(ctx context.Context, user, token string) error
}

// NewService returns a usable service, wrapping a repository.
func NewService(r Repository, v Validator, logger log.Logger) *Service {
	return &Service{
		repo:   r,
		valid:  v,
		logger: logger,
	}
}

// Add a user and their DNA sequence to the database.
func (s *Service) Add(ctx context.Context, user, token, sequence string) (err error) {
	defer func() {
		ctxlog.From(ctx).Log("dna_method", "Add", "add_user", user, "add_err", err)
	}()

	if err := s.valid.Validate(ctx, user, token); err != nil {
		return ErrBadAuth
	}

	if !validSequence(sequence) {
		return ErrInvalidSequence
	}

	if err := s.repo.Insert(ctx, user, sequence); err != nil {
		return errors.Wrap(err, "error adding new user")
	}

	return nil
}

func validSequence(sequence string) bool {
	for _, r := range sequence {
		switch r {
		case 'g', 'a', 't', 'c':
			continue
		default:
			return false
		}
	}
	return true
}

// Check returns true if the given subsequence is present in the user's DNA.
func (s *Service) Check(ctx context.Context, user, token, subsequence string) (err error) {
	defer func() {
		ctxlog.From(ctx).Log("dna_method", "Check", "check_user", user, "check_subseq", subsequence, "check_err", err)
	}()

	if err := s.valid.Validate(ctx, user, token); err != nil {
		return ErrBadAuth
	}

	sequence, err := s.repo.Select(ctx, user)
	if err != nil {
		return errors.Wrap(err, "error reading DNA sequence from repository")
	}

	if !strings.Contains(sequence, subsequence) {
		return ErrSubsequenceNotFound
	}

	return nil
}

// ServeHTTP implements http.Handler in a very naïve way.
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx, ctxlog = ctxlog.New(r.Context(), "http_method", r.Method, "http_path", r.URL.Path)
		first       = extractPathToken(r.URL.Path, 0)
		method      = r.Method
	)
	defer func() { s.logger.Log(ctxlog.Keyvals()...) }()
	switch {
	case method == "POST" && first == "add":
		var (
			user     = r.URL.Query().Get("user")
			token    = r.URL.Query().Get("token")
			sequence = r.URL.Query().Get("sequence")
		)
		err := s.Add(ctx, user, token, sequence)
		switch {
		case err == nil:
			fmt.Fprintln(w, "Add OK")
		case err == ErrBadAuth:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case method == "GET" && first == "check":
		var (
			user        = r.URL.Query().Get("user")
			token       = r.URL.Query().Get("token")
			subsequence = r.URL.Query().Get("subsequence")
		)
		err := s.Check(ctx, user, token, subsequence)
		switch {
		case err == nil:
			fmt.Fprintln(w, "Subsequence found")
		case err == ErrSubsequenceNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case err == ErrBadAuth:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}

func extractPathToken(path string, position int) string {
	toks := strings.Split(strings.Trim(path, "/ "), "/")
	if len(toks) <= position {
		return ""
	}
	return toks[position]
}
