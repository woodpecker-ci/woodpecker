// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func Setup(ctx context.Context, serviceName string) func() {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(version.Version),
	)

	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("failed to setup tracing")
	}

	tracingShutdown, err := registerTraceExporter(res, exp)
	if err != nil {
		// non critical error, just log it.
		log.Warn().Err(err).Msg("failed to setup tracing")
	}

	return func() {
		timeout := 25 * time.Second // nolint:gomnd // 2.5x default OTLP timeout

		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
		defer cancel()

		wg := &sync.WaitGroup{}

		if tracingShutdown != nil {
			wg.Add(1)

			go func() {
				defer wg.Done()

				if err := tracingShutdown(ctx); err != nil {
					log.Error().Err(err).Msg("failed to shutdown tracing")
				}
			}()
		}

		wg.Wait()
	}
}

func registerTraceExporter(res *resource.Resource, exporter sdktrace.SpanExporter) (func(context.Context) error, error) {
	options := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	}
	if res != nil {
		options = append(options, sdktrace.WithResource(res))
	}

	tp := sdktrace.NewTracerProvider(options...)
	otel.SetTracerProvider(tp)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tp.Shutdown, nil
}
