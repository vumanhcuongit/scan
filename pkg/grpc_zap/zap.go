package grpc_zap

import (
	"context"
	"time"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpcRequestId "github.com/infobloxopen/atlas-app-toolkit/requestid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type options struct {
	levelFunc    CodeToLevel
	shouldLog    grpc_logging.Decider
	codeFunc     grpc_logging.ErrorToCode
	durationFunc DurationToField
	messageFunc  MessageProducer
}

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code codes.Code) zapcore.Level

// DurationToField function defines how to produce duration fields for logging
type DurationToField func(duration time.Duration) zapcore.Field

// WithDecider customizes the function for deciding if the gRPC interceptor logs should log.
func WithDecider(f grpc_logging.Decider) Option {
	return func(o *options) {
		o.shouldLog = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f grpc_logging.ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to Zap fields.
func WithDurationField(f DurationToField) Option {
	return func(o *options) {
		o.durationFunc = f
	}
}

// WithMessageProducer customizes the function for message formation.
func WithMessageProducer(f MessageProducer) Option {
	return func(o *options) {
		o.messageFunc = f
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes and interceptor log level for server side.
func DefaultCodeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK:
		return zap.InfoLevel
	case codes.Canceled:
		return zap.InfoLevel
	case codes.Unknown:
		return zap.ErrorLevel
	case codes.InvalidArgument:
		return zap.InfoLevel
	case codes.DeadlineExceeded:
		return zap.WarnLevel
	case codes.NotFound:
		return zap.InfoLevel
	case codes.AlreadyExists:
		return zap.InfoLevel
	case codes.PermissionDenied:
		return zap.WarnLevel
	case codes.Unauthenticated:
		return zap.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return zap.WarnLevel
	case codes.FailedPrecondition:
		return zap.WarnLevel
	case codes.Aborted:
		return zap.WarnLevel
	case codes.OutOfRange:
		return zap.WarnLevel
	case codes.Unimplemented:
		return zap.ErrorLevel
	case codes.Internal:
		return zap.ErrorLevel
	case codes.Unavailable:
		return zap.WarnLevel
	case codes.DataLoss:
		return zap.ErrorLevel
	default:
		return zap.ErrorLevel
	}
}

// DefaultDurationToField is the default implementation of converting request duration to a Zap field.
var DefaultDurationToField = DurationToTimeMillisField

// DurationToTimeMillisField converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisField(duration time.Duration) zapcore.Field {
	return zap.Float32("grpc.time_ms", durationToMilliseconds(duration))
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Zap's encoder settings to determine how that is output.
func DurationToDurationField(duration time.Duration) zapcore.Field {
	return zap.Duration("grpc.duration", duration)
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}

// MessageProducer produces a user defined log message
type MessageProducer func(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field)

// DefaultMessageProducer writes the default message
func DefaultMessageProducer(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field) {
	// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
	ctxzap.Extract(ctx).Check(level, msg).Write(
		zap.Error(err),
		zap.String("grpc.code", code.String()),
		duration,
	)
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		log := ctxzap.Extract(ctx)
		reqID := grpcRequestId.HandleRequestID(ctx)
		reqTime := time.Now()
		ctx = metadata.AppendToOutgoingContext(ctx, "request_id", reqID)
		ctx = metadata.AppendToOutgoingContext(ctx, "Request-Id", reqID)
		log.Sugar().Infow("starting grpc request", "method", method)
		defer func() {
			log.Info("grpc request finished", []zapcore.Field{
				zap.String("method", method),
				zap.Float32("latency", float32(time.Since(reqTime).Nanoseconds()/1000)),
			}...)
		}()

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
