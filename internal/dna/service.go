package dna

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
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
	repo  Repository
	valid Validator
}

// Repository is a client-side interface, which models
// the concrete e.g. SQLiteRepository.
//
// We use an interface instead of the concrete type,
// so we can more easily test the Service.
type Repository interface {
	Insert(user, sequence string) error
	Select(user string) (sequence string, err error)
}

// Validator is a client-side interface, which models
// the parts of the auth service that we use.
type Validator interface {
	Validate(user, token string) error
}

// NewService returns a usable service, wrapping a repository.
func NewService(r Repository, v Validator) *Service {
	return &Service{
		repo:  r,
		valid: v,
	}
}

// Add a user and their DNA sequence to the database.
func (s *Service) Add(user, token, sequence string) error {
	if err := s.valid.Validate(user, token); err != nil {
		return ErrBadAuth
	}
	if !validSequence(sequence) {
		return ErrInvalidSequence
	}
	if err := s.repo.Insert(user, sequence); err != nil {
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
func (s *Service) Check(user, token, subsequence string) error {
	if err := s.valid.Validate(user, token); err != nil {
		return ErrBadAuth
	}
	sequence, err := s.repo.Select(user)
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
		first  = extractPathToken(r.URL.Path, 0)
		method = r.Method
	)
	switch {
	case method == "POST" && first == "add":
		var (
			user     = r.URL.Query().Get("user")
			token    = r.URL.Query().Get("token")
			sequence = r.URL.Query().Get("sequence")
		)
		switch err := s.Add(user, token, sequence); true {
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
		switch err := s.Check(user, token, subsequence); true {
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
