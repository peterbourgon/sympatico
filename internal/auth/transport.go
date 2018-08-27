package auth

import (
	"context"
	"fmt"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewHTTPTransport returns an http.Handler with routes for each endpoint.
// It uses the Go kit style endpoints, and Go kit http.Servers.
func NewHTTPTransport(s *Service) http.Handler {
	r := mux.NewRouter()
	e := makeEndpoints(s)
	{
		r.Methods("POST").Path("/signup").Handler(kithttp.NewServer(e.signup, decodeSignupRequest, encodeSignupResponse))
		r.Methods("POST").Path("/login").Handler(kithttp.NewServer(e.login, decodeLoginRequest, encodeLoginResponse))
		r.Methods("GET").Path("/validate").Handler(kithttp.NewServer(e.validate, decodeValidateRequest, encodeValidateResponse))
		r.Methods("POST").Path("/logout").Handler(kithttp.NewServer(e.logout, decodeLogoutRequest, encodeLogoutResponse))
	}
	return r
}

func decodeSignupRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	return signupRequest{
		user: r.URL.Query().Get("user"),
		pass: r.URL.Query().Get("pass"),
	}, nil
}

func encodeSignupResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(signupResponse)
	switch {
	case resp.err == nil:
		fmt.Fprintln(w, "signup successful")
	case resp.err == ErrBadAuth:
		http.Error(w, resp.err.Error(), http.StatusUnauthorized)
	case resp.err != nil:
		http.Error(w, resp.err.Error(), http.StatusInternalServerError)
	default:
		panic("unreachable")
	}
	return nil
}

func decodeLoginRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	return loginRequest{
		user: r.URL.Query().Get("user"),
		pass: r.URL.Query().Get("pass"),
	}, nil
}

func encodeLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(loginResponse)
	switch {
	case resp.err == nil:
		fmt.Fprintln(w, resp.token)
	case resp.err == ErrBadAuth:
		http.Error(w, resp.err.Error(), http.StatusUnauthorized)
	case resp.err != nil:
		http.Error(w, resp.err.Error(), http.StatusInternalServerError)
	default:
		panic("unreachable")
	}
	return nil
}

func decodeLogoutRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	return logoutRequest{
		user:  r.URL.Query().Get("user"),
		token: r.URL.Query().Get("token"),
	}, nil
}

func encodeLogoutResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(logoutResponse)
	switch {
	case resp.err == nil:
		fmt.Fprintln(w, "logout successful")
	case resp.err == ErrBadAuth:
		http.Error(w, resp.err.Error(), http.StatusUnauthorized)
	case resp.err != nil:
		http.Error(w, resp.err.Error(), http.StatusInternalServerError)
	default:
		panic("unreachable")
	}
	return nil
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	return validateRequest{
		user:  r.URL.Query().Get("user"),
		token: r.URL.Query().Get("token"),
	}, nil
}

func encodeValidateResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(validateResponse)
	switch {
	case resp.err == nil:
		fmt.Fprintln(w, "validate successful")
	case resp.err == ErrBadAuth:
		http.Error(w, resp.err.Error(), http.StatusUnauthorized)
	case resp.err != nil:
		http.Error(w, resp.err.Error(), http.StatusInternalServerError)
	default:
		panic("unreachable")
	}
	return nil
}
