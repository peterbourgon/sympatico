package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/peterbourgon/sympatico/internal/ctxlog"
	"github.com/peterbourgon/sympatico/internal/dna"
	"github.com/peterbourgon/sympatico/internal/usage"
)

func main() {
	fs := flag.NewFlagSet("monolith", flag.ExitOnError)
	var (
		apiAddr     = fs.String("api", "127.0.0.1:8080", "HTTP API listen address")
		authsvcAddr = fs.String("authsvc", "http://127.0.0.1:8081", "HTTP endpoint for authsvc")
		dnaURN      = fs.String("dna-urn", "file:dna.db", "URN for DNA DB")
	)
	fs.Usage = usage.For(fs, "monolith [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var dnaCheckDuration *prometheus.HistogramVec
	{
		dnaCheckDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "dna",
			Name:      "check_duration_seconds",
			Help:      "Time spent performing DNA subsequence checks.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"success"})
	}

	var authsvc dna.Validator
	{
		authsvc = newAuthClient(*authsvcAddr)
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
		r.PathPrefix("/dna/").Handler(http.StripPrefix("/dna", dnaserver))
		api = ctxlog.NewHTTPMiddleware(r, logger)
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
