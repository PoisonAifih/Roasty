package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/poisonaifih/roasty/backend/internal"
	"github.com/poisonaifih/roasty/backend/internal/handlers"
	"github.com/poisonaifih/roasty/backend/internal/services"
)

func main() {
	internal.LoadDotEnv(".env", "../.env")

	databaseURL := env("DATABASE_URL", "postgres://roasty:roasty@127.0.0.1:5433/roasty?sslmode=disable")
	aiURL := env("OPENROUTER_BASE_URL", env("AI_URL", "https://openrouter.ai/api/v1"))
	aiKey := firstNonEmpty(env("OPENROUTER_API_KEY", ""), env("API_KEY", ""))
	aiModel := env("AI_MODEL", "deepseek/deepseek-v4-flash")
	addr := env("ADDR", ":8014")

	ctx := context.Background()
	pool, err := internal.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	if err := internal.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ai := services.NewAIClient(aiURL, aiKey, aiModel)
	scout := services.NewScoutService(pool, ai)
	inv := services.NewInventoryService(pool, ai)
	crm := services.NewCRMService(pool, ai)

	agent := services.NewAgent(pool, ai, scout, inv)

	h := handlers.New(scout, inv, crm, agent, pool)
	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("roasty api listening on %s (model=%s)", addr, aiModel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
