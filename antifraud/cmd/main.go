package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"antifraud/internal/app"
	"antifraud/internal/repository"
	"antifraud/internal/usecases"
	"antifraud/pkg/antifraud"
)

func main() {
	grpcPort := getEnv("GRPC_PORT", "9091")

	db, err := repository.NewCassandraFeatureStore(
		[]string{
			getEnv(
				"CASSANDRA_HOST",
				"127.0.0.1",
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
		sbpChecks     = usecases.NewRuleChecker(nil)
		cashOutChecks = usecases.NewRuleChecker(nil)
		internal      = usecases.NewRuleChecker(nil)
	)

	service := app.NewService(
		&usecases.SbpOutCheck{RuleChecker: sbpChecks},
		&usecases.CashOutCheck{RuleChecker: cashOutChecks},
		&usecases.InternalCheck{RuleChecker: internal},
	)
	server := grpc.NewServer()
	antifraud.RegisterOnlineCheckServer(server, service)
	reflection.Register(server)
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err = server.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
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
