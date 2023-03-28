package main

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "<ChangeMe>",
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		EnableTracing: true,
		Debug:         true,
		TracesSampler: sentry.TracesSampler(func(ctx sentry.SamplingContext) float64 {
			// Don't sample health checks.
			if ctx.Span.Op == "GET /health" {
				return 0.0
			}

			return 1.0
		})}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())

	e := echo.New()
	e.Use(otelecho.Middleware("site-manager"))

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Once it's done, you can attach the handler as one of your middleware
	e.Use(sentryecho.New(sentryecho.Options{}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "Healthy")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
