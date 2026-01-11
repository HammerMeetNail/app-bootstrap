package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/notes-template/internal/config"
	"github.com/example/notes-template/internal/database"
	"github.com/example/notes-template/internal/handlers"
	"github.com/example/notes-template/internal/logging"
	"github.com/example/notes-template/internal/middleware"
	"github.com/example/notes-template/internal/services"
)

func main() {
	if err := run(); err != nil {
		logging.Error("Application error", map[string]interface{}{"error": err.Error()})
		os.Exit(1)
	}
}

func run() error {
	logger := logging.New()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Server.Debug {
		logger.SetLevel(logging.LevelDebug)
		logging.SetDefaultLevel(logging.LevelDebug)
		logger.Debug("Debug logging enabled", map[string]interface{}{
			"max_chars": cfg.Server.DebugMaxChars,
			"env":       cfg.Server.Environment,
		})
	}

	logger.Info("Starting server...")

	logger.Info("Connecting to PostgreSQL", map[string]interface{}{
		"host": cfg.Database.Host,
		"port": cfg.Database.Port,
	})
	db, err := database.NewPostgresDB(cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("connecting to postgres: %w", err)
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL")

	logger.Info("Running database migrations...")
	migrator, err := database.NewMigrator(cfg.Database.DSN(), "migrations")
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	if err := migrator.Up(); err != nil {
		_ = migrator.Close()
		return fmt.Errorf("running migrations: %w", err)
	}
	_ = migrator.Close()
	logger.Info("Migrations completed")

	logger.Info("Connecting to Redis", map[string]interface{}{
		"addr": cfg.Redis.Addr(),
	})
	redisDB, err := database.NewRedisDB(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}
	defer func() { _ = redisDB.Close() }()
	logger.Info("Connected to Redis")

	// Initialize services
	dbAdapter := services.NewPoolAdapter(db.Pool)
	redisAdapter := services.NewRedisAdapter(redisDB.Client)

	userService := services.NewUserService(dbAdapter)
	authService := services.NewAuthService(dbAdapter, redisAdapter)
	emailService := services.NewEmailService(&cfg.Email, dbAdapter)
	noteService := services.NewNoteService(dbAdapter)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db, redisDB)
	authHandler := handlers.NewAuthHandler(userService, authService, emailService, cfg.Server.Secure)
	noteHandler := handlers.NewNoteHandler(noteService)
	pageHandler, err := handlers.NewPageHandler("web/templates")
	if err != nil {
		return fmt.Errorf("loading templates: %w", err)
	}

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService, userService)
	csrfMiddleware := middleware.NewCSRFMiddleware(cfg.Server.Secure)
	securityHeaders := middleware.NewSecurityHeaders(cfg.Server.Secure)
	cacheControl := middleware.NewCacheControl()
	compress := middleware.NewCompress()
	requestLogger := middleware.NewRequestLogger(logger)

	requireAuth := authMiddleware.RequireAuth

	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.HandleFunc("GET /ready", healthHandler.Ready)
	mux.HandleFunc("GET /live", healthHandler.Live)

	// CSRF token endpoint
	mux.Handle("GET /api/csrf", http.HandlerFunc(csrfMiddleware.GetToken))

	// Auth endpoints
	mux.Handle("POST /api/auth/register", http.HandlerFunc(authHandler.Register))
	mux.Handle("POST /api/auth/login", http.HandlerFunc(authHandler.Login))
	mux.Handle("POST /api/auth/logout", requireAuth(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/auth/me", requireAuth(http.HandlerFunc(authHandler.Me)))
	mux.Handle("POST /api/auth/password", requireAuth(http.HandlerFunc(authHandler.ChangePassword)))
	mux.Handle("POST /api/auth/verify-email", http.HandlerFunc(authHandler.VerifyEmail))
	mux.Handle("POST /api/auth/resend-verification", requireAuth(http.HandlerFunc(authHandler.ResendVerification)))
	mux.Handle("POST /api/auth/magic-link", http.HandlerFunc(authHandler.MagicLink))
	mux.Handle("GET /api/auth/magic-link/verify", http.HandlerFunc(authHandler.MagicLinkVerify))
	mux.Handle("POST /api/auth/forgot-password", http.HandlerFunc(authHandler.ForgotPassword))
	mux.Handle("POST /api/auth/reset-password", http.HandlerFunc(authHandler.ResetPassword))

	// Notes endpoints
	mux.Handle("GET /api/notes", requireAuth(http.HandlerFunc(noteHandler.List)))
	mux.Handle("POST /api/notes", requireAuth(http.HandlerFunc(noteHandler.Create)))
	mux.Handle("GET /api/notes/{id}", requireAuth(http.HandlerFunc(noteHandler.Get)))
	mux.Handle("PUT /api/notes/{id}", requireAuth(http.HandlerFunc(noteHandler.Update)))
	mux.Handle("DELETE /api/notes/{id}", requireAuth(http.HandlerFunc(noteHandler.Delete)))

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// API Docs redirect
	mux.Handle("GET /api/docs", http.RedirectHandler("/static/swagger/index.html", http.StatusFound))

	// SPA route - serve index.html for the root path
	mux.Handle("GET /{$}", http.HandlerFunc(pageHandler.Index))

	// Build middleware chain (order matters: outermost first)
	var handler http.Handler = mux
	handler = authMiddleware.Authenticate(handler)
	handler = csrfMiddleware.Protect(handler)
	handler = cacheControl.Apply(handler)
	handler = compress.Apply(handler)
	handler = securityHeaders.Apply(handler)
	handler = requestLogger.Apply(handler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan bool, 1)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shutdown the server", map[string]interface{}{
				"error": err.Error(),
			})
		}
		close(done)
	}()

	logger.Info("Server listening", map[string]interface{}{
		"addr": addr,
	})
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	logger.Info("Server stopped")
	return nil
}
