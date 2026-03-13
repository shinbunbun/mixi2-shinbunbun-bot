package main

import (
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mixigroup/mixi2-application-sdk-go/auth"
	"github.com/mixigroup/mixi2-application-sdk-go/event/webhook"
	application_apiv1 "github.com/mixigroup/mixi2-application-sdk-go/gen/go/social/mixi/application/service/application_api/v1"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/config"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/github"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/handler"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/mixi2"
	"github.com/shinbunbun/mixi2-shinbunbun-bot/internal/scheduler"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Decode Ed25519 public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(cfg.SignaturePublicKey)
	if err != nil {
		log.Fatalf("failed to decode public key: %v", err)
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)

	// Create authenticator
	authenticator, err := auth.NewAuthenticator(cfg.ClientID, cfg.ClientSecret, cfg.TokenURL)
	if err != nil {
		log.Fatalf("failed to create authenticator: %v", err)
	}

	// Create gRPC connection
	apiConn, err := grpc.NewClient(
		cfg.APIAddress,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
	)
	if err != nil {
		log.Fatalf("failed to connect to api: %v", err)
	}
	defer apiConn.Close()

	apiClient := application_apiv1.NewApplicationServiceClient(apiConn)

	// Create mixi2 client
	mixi2Client := mixi2.NewClient(apiClient, authenticator)

	// Create GitHub client
	githubClient := github.NewClient()

	// Start scheduler
	sched := scheduler.New(githubClient, mixi2Client)
	if err := sched.Start(cfg.DailyPostCron); err != nil {
		log.Fatalf("failed to start scheduler: %v", err)
	}
	defer sched.Stop()

	// Create webhook handler
	webhookHandler := handler.NewWebhookHandler()

	// Start webhook server
	webhookAddr := ":" + cfg.Port
	webhookServer := webhook.NewServer(webhookAddr, publicKey, webhookHandler, webhook.WithLogger(logger))

	// Start health check server
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	healthServer := &http.Server{
		Addr:    ":" + cfg.HealthPort,
		Handler: healthMux,
	}
	go func() {
		logger.Info("starting health check server", slog.String("port", cfg.HealthPort))
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("health server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		logger.Info("shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := webhookServer.Shutdown(ctx); err != nil {
			logger.Error("webhook server shutdown error", slog.Any("error", err))
		}
		if err := healthServer.Shutdown(ctx); err != nil {
			logger.Error("health server shutdown error", slog.Any("error", err))
		}
	}()

	// Start webhook server (blocking)
	logger.Info("starting webhook server", slog.String("port", cfg.Port))
	if err := webhookServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("webhook server error: %v", err)
	}
	logger.Info("stopped")
}
