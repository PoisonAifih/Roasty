package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	qdrantURL := env("QDRANT_URL", "http://localhost:6333")
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
	rag := services.NewRAGService(qdrantURL, ai)
	scout := services.NewScoutService(pool, ai)
	inv := services.NewInventoryService(pool, ai)
	crm := services.NewCRMService(pool, ai)
	notif := services.NewNotificationService(pool, inv, crm)

	agent := services.NewAgent(pool, ai, scout, inv, crm, rag)

	h := handlers.New(scout, inv, crm, agent, notif, pool)
	mux := http.NewServeMux()
	h.Register(mux)

	srv := &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Notification cron — runs Check on startup then every NOTIF_INTERVAL_MIN.
	cronCtx, cronCancel := context.WithCancel(ctx)
	go func() {
		intervalMin := envInt("NOTIF_INTERVAL_MIN", 60)
		log.Printf("notif cron starting (interval=%dm)", intervalMin)
		if err := notif.Check(cronCtx); err != nil {
			log.Printf("notif check: %v", err)
		}
		ticker := time.NewTicker(time.Duration(intervalMin) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := notif.Check(cronCtx); err != nil {
					log.Printf("notif check: %v", err)
				}
			case <-cronCtx.Done():
				log.Println("notif cron stopped")
				return
			}
		}
	}()

	go func() {
		log.Printf("roasty api listening on %s (model=%s)", addr, aiModel)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	cronCancel()
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

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
