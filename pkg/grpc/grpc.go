package grpc

import (
	"crypto/tls"
	"net/url"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	otgrpc "github.com/opentracing-contrib/go-grpc"
	ot "github.com/opentracing/opentracing-go"
	"github.com/vumanhcuongit/scan/internal/config"
	grpcZap "github.com/vumanhcuongit/scan/pkg/grpc_zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

// CreateGRPCServer ...
func CreateGRPCServer(cfg *config.EnvConfig) *grpc.Server {
	grpc_prometheus.EnableHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.5, 0.9, 0.99, 1.0}),
	)

	return grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
		)),
	)
}

// CreateGRPCClientConn ...
func CreateGRPCClientConn(host string, tlsEnabled bool) (*grpc.ClientConn, error) {
	// nolint
	secureOption := grpc.WithInsecure()
	if tlsEnabled {
		creds := credentials.NewTLS(nil)
		secureOption = grpc.WithTransportCredentials(creds)
	}

	grpc_prometheus.EnableClientHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.5, 0.9, 0.99, 1.0}),
	)

	return grpc.Dial(
		host,
		secureOption,
		grpc.WithChainUnaryInterceptor(
			grpcZap.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		),
	)
}

// CreateGRPCClientConnFromDSN ...
func CreateGRPCClientConnFromDSN(dsn string) (*grpc.ClientConn, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	// nolint
	secureOption := grpc.WithInsecure()
	sslmode := q.Get("sslmode")
	switch sslmode {
	case "true":
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		secureOption = grpc.WithTransportCredentials(creds)
	case "require":
		creds := credentials.NewTLS(&tls.Config{})
		secureOption = grpc.WithTransportCredentials(creds)
	}

	gzipOption := grpc.WithDefaultCallOptions()
	gzipmode := q.Get("gzip")
	switch gzipmode {
	case "true":
		gzipOption = grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name))
	}

	grpc_prometheus.EnableClientHandlingTimeHistogram(
		grpc_prometheus.WithHistogramBuckets([]float64{0.5, 0.9, 0.99, 1.0}),
	)

	return grpc.Dial(
		u.Host,
		secureOption,
		gzipOption,
		grpc.WithChainUnaryInterceptor(
			grpcZap.UnaryClientInterceptor(),
			otgrpc.OpenTracingClientInterceptor(
				ot.GlobalTracer(),
				otgrpc.IncludingSpans(func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
					return method != "/grpc.health.v1.Health/Check"
				}),
			),
			grpc_prometheus.UnaryClientInterceptor,
		),
	)
}
