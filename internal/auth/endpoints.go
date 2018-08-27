package auth

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/ratelimit"
	"golang.org/x/time/rate"
)

// endpoints collects all of the endpoints that compose an authsvc.
// It's meant to be used as a helper struct, collecting all the
// individual elements into a single parameter.
type endpoints struct {
	signup   endpoint.Endpoint
	login    endpoint.Endpoint
	validate endpoint.Endpoint
	logout   endpoint.Endpoint
}

func makeEndpoints(s *Service) endpoints {
	var (
		limitSignup = ratelimit.NewErroringLimiter(rate.NewLimiter(3, 3)) // 3 QPS
		limitLogin  = ratelimit.NewDelayingLimiter(rate.NewLimiter(1, 3)) // 1 QPS
	)
	return endpoints{
		signup:   limitSignup(makeSignupEndpoint(s)),
		login:    limitLogin(makeLoginEndpoint(s)),
		validate: makeValidateEndpoint(s),
		logout:   makeLogoutEndpoint(s),
	}
}

func makeSignupEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(signupRequest)
		serr := s.Signup(ctx, req.user, req.pass)
		return signupResponse{err: serr}, nil
	}
}

type signupRequest struct {
	user string
	pass string
}

type signupResponse struct {
	err error
}

func makeLoginEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(loginRequest)
		token, serr := s.Login(ctx, req.user, req.pass)
		return loginResponse{token: token, err: serr}, nil
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

func makeLogoutEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(logoutRequest)
		serr := s.Logout(ctx, req.user, req.token)
		return logoutResponse{err: serr}, nil
	}
}

type logoutRequest struct {
	user  string
	token string
}

type logoutResponse struct {
	err error
}

func makeValidateEndpoint(s *Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateRequest)
		serr := s.Validate(ctx, req.user, req.token)
		return validateResponse{err: serr}, nil
	}
}

type validateRequest struct {
	user  string
	token string
}

type validateResponse struct {
	err error
}
