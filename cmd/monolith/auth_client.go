package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/peterbourgon/sympatico/internal/dna"
)

func newAuthClient(addr string) dna.Validator {
	return authClient(addr)
}

type authClient string // base URL

func (c authClient) Validate(ctx context.Context, user, token string) error {
	url := fmt.Sprintf("%s/auth/validate?user=%s&token=%s", c, user, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.Wrap(err, "error constructing validate request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error making validate request")
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("authsvc returned %s", resp.Status)
	}
	return nil
}
