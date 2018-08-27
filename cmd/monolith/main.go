package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/pkg/errors"

	"github.com/peterbourgon/sympatico/internal/auth"
	"github.com/peterbourgon/sympatico/internal/dna"
)

func main() {
	fs := flag.NewFlagSet("monolith", flag.ExitOnError)
	var (
		apiAddr = fs.String("api", "127.0.0.1:8080", "HTTP API listen address")
		authURN = fs.String("auth-urn", "file:auth.db", "URN for auth DB")
		dnaURN  = fs.String("dna-urn", "file:dna.db", "URN for DNA DB")
	)
	fs.Usage = usageFor(fs, "monolith [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var authsvc *auth.Service
	{
		authrepo, err := auth.NewSQLiteRepository(*authURN)
		if err != nil {
			logger.Log("during", "auth.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		authsvc = auth.NewService(authrepo)
	}

	var dnasvc *dna.Service
	{
		dnarepo, err := dna.NewSQLiteRepository(*dnaURN)
		if err != nil {
			logger.Log("during", "dna.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		dnasvc = dna.NewService(dnarepo, authsvc, logger)
	}

	var api http.Handler
	{
		// The HTTP API mounts endpoints to be consumed by clients.
		r := mux.NewRouter()

		// One way to make a service accessible over HTTP is to write individual
		// handle functions that translate to and from HTTP semantics. Note that
		// we don't bind the auth validate method, because that's only used by
		// other components, never by clients directly.
		r.Methods("POST").Path("/auth/signup").HandlerFunc(handleSignup(authsvc, logger))
		r.Methods("POST").Path("/auth/login").HandlerFunc(handleLogin(authsvc, logger))
		r.Methods("POST").Path("/auth/logout").HandlerFunc(handleLogout(authsvc, logger))

		// Another way to make a service accessible over HTTP is to have the
		// service implement http.Handler directly, via a ServeHTTP method.
		r.PathPrefix("/dna/").Handler(http.StripPrefix("/dna", dnasvc))

		api = r
	}

	var g run.Group
	{
		server := &http.Server{
			Addr:    *apiAddr,
			Handler: api,
		}
		g.Add(func() error {
			logger.Log("component", "API", "addr", *apiAddr)
			return server.ListenAndServe()
		}, func(error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			server.Shutdown(ctx)
		})
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-c:
				return errors.Errorf("received signal %s", sig)
			}
		}, func(error) {
			cancel()
		})
	}
	logger.Log("exit", g.Run())
}

func handleSignup(s *auth.Service, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user = r.URL.Query().Get("user")
			pass = r.URL.Query().Get("pass")
		)
		err := s.Signup(r.Context(), user, pass)
		if err == auth.ErrBadAuth {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		fmt.Fprintln(w, "signup OK")
		logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", nil)
	}
}

func handleLogin(s *auth.Service, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user = r.URL.Query().Get("user")
			pass = r.URL.Query().Get("pass")
		)
		token, err := s.Login(r.Context(), user, pass)
		if err == auth.ErrBadAuth {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		fmt.Fprintln(w, token)
		logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", nil)
	}
}

func handleLogout(s *auth.Service, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user  = r.URL.Query().Get("user")
			token = r.URL.Query().Get("token")
		)
		err := s.Logout(r.Context(), user, token)
		if err == auth.ErrBadAuth {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", err)
			return
		}
		fmt.Fprintln(w, "logout OK")
		logger.Log("http_method", r.Method, "http_path", r.URL.Path, "user", user, "err", nil)
	}
}

func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stdout, "USAGE\n")
		fmt.Fprintf(os.Stdout, "  %s\n", short)
		fmt.Fprintf(os.Stdout, "\n")
		fmt.Fprintf(os.Stdout, "FLAGS\n")
		tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			def := f.DefValue
			if def == "" {
				def = "..."
			}
			fmt.Fprintf(tw, "  -%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		tw.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
