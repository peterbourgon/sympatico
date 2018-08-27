package auth

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestFlow(t *testing.T) {
	s := NewService(newMockRepo())

	if want, have := error(nil), s.Signup("peter", "123456"); want != have {
		t.Fatalf("Signup: want %v, have %v", want, have)
	}

	token, err := s.Login("peter", "123456")
	if want, have := error(nil), err; want != have {
		t.Fatalf("Login: want %v, have %v", want, have)
	}

	if want, have := error(nil), s.Validate("peter", token); want != have {
		t.Errorf("Validate: want %v, have %v", want, have)
	}

	if want, have := error(nil), s.Logout("peter", token); want != have {
		t.Errorf("Logout: want %v, have %v", want, have)
	}

	if want, have := ErrBadAuth, s.Validate("peter", token); want != have {
		t.Errorf("Validate after Logout: want %v, have %v", want, have)
	}
}

type mockRepo struct {
	creds  map[string]string
	tokens map[string]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		creds:  map[string]string{},
		tokens: map[string]string{},
	}
}

func (r *mockRepo) Create(user, pass string) error {
	if _, ok := r.creds[user]; ok {
		return errors.New("user already exists")
	}

	r.creds[user] = pass
	return nil
}

func (r *mockRepo) Auth(user, pass string) (token string, err error) {
	if have, ok := r.creds[user]; !ok || pass != have {
		return "", ErrBadAuth
	}

	p := make([]byte, 8)
	rand.New(rand.NewSource(time.Now().UnixNano())).Read(p)
	token = fmt.Sprintf("%x", p)
	r.tokens[user] = token
	return token, nil
}

func (r *mockRepo) Deauth(user, token string) error {
	if have, ok := r.tokens[user]; !ok || token != have {
		return ErrBadAuth
	}
	delete(r.tokens, user)
	return nil
}

func (r *mockRepo) Validate(user, token string) error {
	if have, ok := r.tokens[user]; !ok || token != have {
		return ErrBadAuth
	}

	return nil
}
