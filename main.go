package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
)

const serviceName = "mkm-test"

func foo(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "foo")
	defer span.Finish()

	bucketID := uint64(0x3c0bd4c89186ca89)

	span.LogFields(
		log.String("foo", "bar"),
		log.Uint64("bucket-id", bucketID),
		log.String("bucket-id-str", fmt.Sprintf("%x", bucketID)),
	)
}

func reporter() (jaeger.Reporter, error) {
	var reporters []jaeger.Reporter

	reporters = append(reporters, jaeger.NewLoggingReporter(jaeger.StdLogger))

	agentHost := "localhost"
	if agentHost != "" {
		var agentHostPort string
		if agentPortStr := os.Getenv("JAEGER_AGENT_PORT"); agentPortStr == "" {
			agentHostPort = fmt.Sprintf("%s:%d", agentHost, jaeger.DefaultUDPSpanServerPort)
		} else {
			agentHostPort = fmt.Sprintf("%s:%s", agentHost, agentPortStr)
		}

		sender, err := jaeger.NewUDPTransport(agentHostPort, 0)
		if err != nil {
			return nil, err
		}
		reporter := jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.Logger(jaeger.StdLogger))
		reporters = append(reporters, reporter)
	}

	return jaeger.NewCompositeReporter(reporters...), nil
}

func InitTracing() io.Closer {
	reporter, err := reporter()
	if err != nil {
		panic(err)
	}

	tracer, closer := jaeger.NewTracer(
		serviceName,
		jaeger.NewConstSampler(true),
		reporter)

	opentracing.SetGlobalTracer(tracer)

	span := opentracing.GlobalTracer().StartSpan("init-test")
	span.Finish()

	return closer
}

func mainE() error {
	defer InitTracing().Close()

	ctx := context.Background()
	foo(ctx)

	return nil
}

func main() {
	if err := mainE(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
