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

	var authEventsTotal *prometheus.CounterVec
	var dnaCheckDuration *prometheus.HistogramVec
	{
		authEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "auth",
			Name:      "events_total",
			Help:      "Total number of auth events.",
		}, []string{"method", "success"})
		dnaCheckDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "dna",
			Name:      "check_duration_seconds",
			Help:      "Time spent performing DNA subsequence checks.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"success"})
	}

	var authsvc *auth.Service
	{
		authrepo, err := auth.NewSQLiteRepository(*authURN)
		if err != nil {
			logger.Log("during", "auth.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		authsvc = auth.NewService(authrepo, authEventsTotal)
	}

	var authserver *auth.HTTPServer
	{
		authserver = auth.NewHTTPServer(authsvc)
	}

	var dnasvc *dna.Service
	{
		dnarepo, err := dna.NewSQLiteRepository(*dnaURN)
		if err != nil {
			logger.Log("during", "dna.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		dnasvc = dna.NewService(dnarepo, authsvc, dnaCheckDuration)
	}

	var dnaserver *dna.HTTPServer
	{
		dnaserver = dna.NewHTTPServer(dnasvc)
	}

	var api http.Handler
	{
		r := mux.NewRouter()
		r.PathPrefix("/auth/").Handler(http.StripPrefix("/auth", authserver))
		r.PathPrefix("/dna/").Handler(http.StripPrefix("/dna", dnaserver))
		api = newLoggingMiddleware(r, logger)
	}

	var g run.Group
	{
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.Handle("/", api)
		server := &http.Server{
			Addr:    *apiAddr,
			Handler: mux,
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
