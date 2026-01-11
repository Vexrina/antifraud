package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"antifraud/internal/app"
	"antifraud/internal/outgoing"
	"antifraud/internal/repository"
	"antifraud/internal/usecases"
	"antifraud/pkg/antifraud"
)

func main() {
	grpcPort := getEnv("GRPC_PORT", "9091")

	// cassandra
	db, err := repository.NewCassandraFeatureStore(
		[]string{
			getEnv(
				"CASSANDRA_HOST",
				"localhost",
			),
		},
		getEnv(
			"CASSANDRA_KEYSPACE",
			"antifraud",
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var (
		// sbp rules
		sbpChecks = usecases.NewRuleChecker(
			[]usecases.Rule{
				outgoing.NewBigSpent(
					true,
					75_000_00,
					90_000_00,
					db,
				),
				outgoing.NewManyPartners(
					true,
					10_000_00,
					10,
					1,
					db,
				),
			},
		)
		// cash out rules
		cashOutChecks = usecases.NewRuleChecker(
			[]usecases.Rule{
				outgoing.NewBigSpent(
					true,
					75_000_00,
					90_000_00,
					db,
				),
				outgoing.NewLargeCashOutDuring1H(
					true,
					45_000_00,
					110_000_00,
					db,
				),
			})
		// internal rules
		internal = usecases.NewRuleChecker(
			[]usecases.Rule{
				outgoing.NewBigSpent(
					true,
					75_000_00,
					90_000_00,
					db,
				),
				outgoing.NewManyPartners(
					true,
					10_000_00,
					10,
					0,
					db,
				),
			},
		)
	)

	// metrics
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	grpcRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "antifraud",
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total gRPC requests",
		},
		[]string{"method", "code"},
	)

	grpcLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "antifraud",
			Subsystem: "grpc",
			Name:      "request_duration_seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)
	grpcErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "antifraud",
			Subsystem: "grpc",
			Name:      "errors_total",
			Help:      "Total gRPC errors by method and type",
		},
		[]string{"method", "reason"}, // reason = limit_overflow, declined_by_af, context_canceled, etc
	)

	reg.MustRegister(grpcRequests, grpcLatency, grpcErrors, usecases.FraudErrors)

	// server
	service := app.NewService(
		&usecases.SbpOutCheck{RuleChecker: sbpChecks},
		&usecases.InternalCheck{RuleChecker: internal},
		&usecases.CashOutCheck{RuleChecker: cashOutChecks},
	)
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			metricsUnaryInterceptor(grpcRequests, grpcLatency, grpcErrors),
		),
	)
	antifraud.RegisterOnlineCheckServer(server, service)
	reflection.Register(server)
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// run grpc
	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err = server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// run metrics
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
		log.Println("metrics listening on :9101")
		_ = http.ListenAndServe(":9101", nil)
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")
	server.GracefulStop()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
