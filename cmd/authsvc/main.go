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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/peterbourgon/sympatico/internal/auth"
)

func main() {
	fs := flag.NewFlagSet("authsvc", flag.ExitOnError)
	var (
		apiAddr = fs.String("api", "127.0.0.1:8081", "HTTP API listen address")
		authURN = fs.String("auth-urn", "auth.db", "URN for auth DB")
	)
	fs.Usage = usageFor(fs, "authsvc [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var authEventsTotal *prometheus.CounterVec
	{
		authEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "auth",
			Name:      "events_total",
			Help:      "Total number of auth events.",
		}, []string{"method", "success"})
	}

	var authrepo *auth.SQLiteRepository
	{
		var err error
		authrepo, err = auth.NewSQLiteRepository(*authURN)
		if err != nil {
			logger.Log("during", "auth.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
	}

	var authsvc *auth.Service
	{
		authsvc = auth.NewService(authrepo, authEventsTotal)
	}

	var authserver *auth.HTTPServer
	{
		authserver = auth.NewHTTPServer(authsvc)
	}

	var api http.Handler
	{
		r := mux.NewRouter()
		r.PathPrefix("/auth/").Handler(http.StripPrefix("/auth", authserver))
		api = newLoggingMiddleware(r, logger)
	}

	var g run.Group
	{
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.Handle("/", api)
		server := &http.Server{Addr: *apiAddr, Handler: mux}
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
