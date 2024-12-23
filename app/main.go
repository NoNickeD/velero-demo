package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Method     string                 `json:"method"`
	API        string                 `json:"api"`
	StatusCode int                    `json:"status_code"`
	Fields     map[string]interface{} `json:"fields"`
}

const (
	numLogEntries = 10
)

var (
	tracer trace.Tracer
	logger *zap.Logger
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// Set up logger
	var err error
	logger, err = setupLogger()
	if err != nil {
		log.Fatalf("failed to initialize zap logger: %v", err)
	}
	defer logger.Sync()

	// Set up OpenTelemetry Tracing
	tracerProvider, err := setupTracer()
	if err != nil {
		logger.Fatal("failed to initialize trace provider", zap.Error(err))
	}
	otel.SetTracerProvider(tracerProvider)
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("failed to shutdown TracerProvider", zap.Error(err))
		}
	}()

	// Set global tracer
	tracer = otel.Tracer("srekubecraftio-tracer")

	// Initialize server and routes
	handler := http.NewServeMux()
	handler.Handle("/generate", otelhttp.NewHandler(http.HandlerFunc(generateLogsHandler), "GenerateLogs"))
	handler.HandleFunc("/healthz", healthCheckHandler) // Health check endpoint

	srv := &http.Server{
		Addr:    getServerAddress(),
		Handler: middleware(handler),
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		logger.Info("Starting server", zap.String("port", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}

func setupLogger() (*zap.Logger, error) {
	return zap.NewProduction()
}

func setupTracer() (*sdktrace.TracerProvider, error) {
	// Retrieve OTLP endpoint from environment variable
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "tempo.monitoring:4317" // default to Tempo endpoint if not set
	}

	// Set up the OTLP trace exporter using gRPC
	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // Block until the connection is established
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OTLP trace exporter: %w", err)
	}

	// Define resource attributes for tracing
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", "srekubecraftio-service"),
			attribute.String("service.version", "1.0.0"),
			attribute.String("deployment.environment", "development"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create and return a TracerProvider with the OTLP trace exporter
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	), nil
}

func getServerAddress() string {
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		return addr
	}
	return ":8080"
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func generateLogsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	ctx, span := tracer.Start(ctx, "generateLogsHandler", trace.WithAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.Path),
	))
	defer span.End()

	logEntries := generateLogEntries(ctx, r.Method)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(logEntries); err != nil {
		span.SetStatus(codes.Error, "failed to encode response")
		span.RecordError(err)
		logger.Error("failed to encode response", zap.Error(err))
	}
	span.AddEvent("Finished generating log entries")
}

func generateLogEntries(ctx context.Context, _ string) []LogEntry {
	logEntries := make([]LogEntry, numLogEntries)
	for i := 0; i < numLogEntries; i++ {
		_, logSpan := tracer.Start(ctx, fmt.Sprintf("createLogEntry-%d", i+1))
		duration := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(duration) // Simulate some processing time

		statusCode := getRandomStatusCode()
		level := getLogLevel(statusCode)

		entry := LogEntry{
			Timestamp:  time.Now().Format(time.RFC3339),
			Level:      level,
			Message:    fmt.Sprintf("Log entry %d", i+1),
			Method:     getRandomLogMethod(),
			StatusCode: statusCode,
			API:        getRandomAPI(),
			Fields: map[string]interface{}{
				"request_id": generateRequestID(),
				"iteration":  i + 1,
				"random":     rand.Intn(100),
			},
		}
		logEntries[i] = entry
		logStructuredEntry(entry)
		logSpan.End()
	}
	return logEntries
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getLogLevel(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "ERROR"
	case statusCode >= 400:
		return "WARN"
	case statusCode >= 200 && statusCode < 300:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func getRandomLogMethod() string {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	return methods[rand.Intn(len(methods))]
}

func getRandomStatusCode() int {
	statusCodes := []int{200, 400, 404, 500, 503}
	return statusCodes[rand.Intn(len(statusCodes))]
}

func getRandomAPI() string {
	apis := []string{"/api/v1/users", "/api/v1/products", "/api/v1/orders", "/api/v1/inventory"}
	return apis[rand.Intn(len(apis))]
}

func generateRequestID() string {
	return uuid.New().String()
}

func logStructuredEntry(entry LogEntry) {
	logger.Info("Generated log entry",
		zap.String("timestamp", entry.Timestamp),
		zap.String("level", entry.Level),
		zap.String("message", entry.Message),
		zap.String("method", entry.Method),
		zap.String("api", entry.API),
		zap.Int("status_code", entry.StatusCode),
		zap.Any("fields", entry.Fields),
	)
}
