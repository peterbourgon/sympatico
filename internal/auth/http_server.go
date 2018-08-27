package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NewGoKitHandler returns an http.Handler with routes for each endpoint.
// It uses the Go kit style endpoints, and Go kit http.Servers.
func NewGoKitHandler(service *Service) http.Handler {
	r := mux.NewRouter()
	{
		r.Methods("POST").Path("/signup").Handler(kithttp.NewServer(makeSignupEndpoint(service), decodeSignupRequest, encodeSignupResponse))
		r.Methods("POST").Path("/login").Handler(kithttp.NewServer(makeLoginEndpoint(service), decodeLoginRequest, encodeLoginResponse))
		r.Methods("GET").Path("/validate").Handler(kithttp.NewServer(makeValidateEndpoint(service), decodeValidateRequest, encodeValidateResponse))
		r.Methods("POST").Path("/logout").Handler(kithttp.NewServer(makeLogoutEndpoint(service), decodeLogoutRequest, encodeLogoutResponse))
	}
	return r
}

type signupRequest struct {
	user string
	pass string
}

type signupResponse struct {
	err error
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

func makeSignupEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(signupRequest)
		serr := s.Signup(ctx, req.user, req.pass)
		return signupResponse{err: serr}, nil
	}
}

type loginRequest struct {
	user string
	pass string
}

type loginResponse struct {
	token string
	err   error
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

func makeLoginEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(loginRequest)
		token, serr := s.Login(ctx, req.user, req.pass)
		return loginResponse{token: token, err: serr}, nil
	}
}

type logoutRequest struct {
	user  string
	token string
}

type logoutResponse struct {
	err error
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

func makeLogoutEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(logoutRequest)
		serr := s.Logout(ctx, req.user, req.token)
		return logoutResponse{err: serr}, nil
	}
}

type validateRequest struct {
	user  string
	token string
}

type validateResponse struct {
	err error
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

func makeValidateEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateRequest)
		serr := s.Validate(ctx, req.user, req.token)
		return validateResponse{err: serr}, nil
	}
}
