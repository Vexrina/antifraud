package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func metricsUnaryInterceptor(
	reqs *prometheus.CounterVec,
	lat *prometheus.HistogramVec,
	errs *prometheus.CounterVec,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		start := time.Now()
		resp, err := handler(ctx, req)
		elapsed := time.Since(start).Seconds()

		method := info.FullMethod
		code := status.Code(err).String()

		// общий счётчик запросов
		reqs.WithLabelValues(method, code).Inc()

		// latency
		lat.WithLabelValues(method).Observe(elapsed)

		// ошибки
		if err != nil {
			reason := categorizeError(err)
			errs.WithLabelValues(method, reason).Inc()
		}

		return resp, err
	}
}

func categorizeError(err error) string {
	switch {
	case errors.Is(err, context.Canceled):
		return "context_canceled"
	case strings.Contains(err.Error(), "declined"):
		return "fraud_operation"
	default:
		return "other"
	}
}
