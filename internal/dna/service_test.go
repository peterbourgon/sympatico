package dna

import (
	"testing"

	"github.com/pkg/errors"
)

func TestFlow(t *testing.T) {
	var (
		repo  = newMockRepo()
		user  = "vincent"
		token = "some_token"
		valid = newMockValidator(user, token)
		s     = NewService(repo, valid)
	)

	if want, have := ErrBadAuth, s.Add(user, "invalid_token", "gattaca"); want != have {
		t.Errorf("Add with bad token: want %v, have %v", want, have)
	}
	if want, have := error(nil), s.Add("vincent", "some_token", "gattaca"); want != have {
		t.Errorf("Add: want %v, have %v", want, have)
	}

	for subsequence, want := range map[string]error{
		"":         nil,
		"g":        nil,
		"ga":       nil,
		"gattac":   nil,
		"gattaca":  nil,
		"x":        ErrSubsequenceNotFound,
		"gata":     ErrSubsequenceNotFound,
		"gattacaa": ErrSubsequenceNotFound,
	} {
		if have := s.Check("vincent", "some_token", subsequence); want != have {
			t.Errorf("Check(%q): want %v, have %v", subsequence, want, have)
		}
	}
}

type mockRepo struct {
	dna map[string]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		dna: map[string]string{},
	}
}

func (r *mockRepo) Insert(user, sequence string) error {
	if _, ok := r.dna[user]; ok {
		return errors.New("user already exists")
	}

	r.dna[user] = sequence
	return nil
}

func (r *mockRepo) Select(user string) (sequence string, err error) {
	sequence, ok := r.dna[user]
	if !ok {
		return "", ErrInvalidUser
	}
	return sequence, nil
}

type mockValidator struct {
	tokens map[string]string
}

func newMockValidator(usertokens ...string) *mockValidator {
	tokens := map[string]string{}
	for i := 0; i < len(usertokens); i += 2 {
		tokens[usertokens[i]] = usertokens[i+1]
	}
	return &mockValidator{
		tokens: tokens,
	}
}

func (v *mockValidator) Validate(user, token string) error {
	if have, ok := v.tokens[user]; !ok || token != have {
		return ErrBadAuth
	}
	return nil
}
