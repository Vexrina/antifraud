package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"processing_core/internal/app"
	"processing_core/internal/helpers"
	"processing_core/internal/outbox"
	"processing_core/internal/producer"
	"processing_core/internal/repository"
	"processing_core/internal/usecases"
	"processing_core/pkg/antifraud"
	"processing_core/pkg/core"
)

type Publisher interface {
	Run(ctx context.Context) error
	Publish(ctx context.Context) error
}

func main() {
	ctx := context.Background()
	grpcPort := getEnv("GRPC_PORT", "8081")

	db := repository.NewDB(
		ctx,
		getEnv(
			"DB_CONNSTR",
			"postgres://proc_core_user:proc_core_pwd@localhost:5433/proc_core_db?sslmode=disable",
		),
	)
	pr, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": getEnv(
				"KAFKA_PRODUCER",
				"localhost:19092",
			),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	afConn := getConn(
		getEnv("AF_API_HOST", "localhost"),
		getEnv("AF_API_PORT", "8082"),
	)
	defer afConn.Close()
	afChecker := antifraud.NewOnlineCheckClient(afConn)

	var (
		afHelper       = helpers.NewAfIntegration(afChecker)
		sbpIntegration = helpers.NewSbpIntegration()
		atmIntegration = helpers.NewAtmIntegration()

		sbpUsecase      = usecases.NewSbpOutgoingUsecase(db, db, afHelper, sbpIntegration)
		internalUsecase = usecases.NewInternalUsecase(db, db, afHelper)
		cashInUsecase   = usecases.NewCashInUsecase(db, db)
		cashOutUsecase  = usecases.NewCashOutUsecase(db, db, afHelper, atmIntegration)

		transactionPublisher = outbox.NewKafkaOutboxPublisher(db, db, producer.NewKafkaProducer(pr, lo.ToPtr(outbox.TransactionTopic)))
	)

	service := app.NewService(
		sbpUsecase,
		internalUsecase,
		cashInUsecase,
		cashOutUsecase,
	)
	server := grpc.NewServer()
	core.RegisterCoreServer(server, service)
	reflection.Register(server)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		_ = transactionPublisher.Run(ctx)
	}()

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

func getConn(url, port string) *grpc.ClientConn {
	log.Println("URL:", url)
	log.Println("PORT:", port)
	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", url, port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Не создали коннекшн с floatweaver: %v", err)
	}
	return conn
}
