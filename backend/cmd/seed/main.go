// seed populates Qdrant with the Roasty knowledge base.
// Run from the backend directory:
//
//	go run ./cmd/seed
//
// The command is idempotent — re-running overwrites existing points with the
// same IDs, so it is safe to re-seed after adding new chunks to ragdocs.go.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/poisonaifih/roasty/backend/internal"
	"github.com/poisonaifih/roasty/backend/internal/services"
)

func main() {
	internal.LoadDotEnv(".env", "../.env")

	qdrantURL := env("QDRANT_URL", "http://localhost:6333")
	aiURL := env("OPENROUTER_BASE_URL", env("AI_URL", "https://openrouter.ai/api/v1"))
	aiKey := firstNonEmpty(env("OPENROUTER_API_KEY", ""), env("API_KEY", ""))
	aiModel := env("AI_MODEL", "deepseek/deepseek-v4-flash")

	if aiKey == "" {
		log.Fatal("OPENROUTER_API_KEY is required for embedding")
	}

	ai := services.NewAIClient(aiURL, aiKey, aiModel)
	rag := services.NewRAGService(qdrantURL, ai)
	ctx := context.Background()

	log.Printf("target qdrant: %s", qdrantURL)
	log.Println("ensuring collection…")
	if err := rag.EnsureCollection(ctx); err != nil {
		log.Fatalf("ensure collection: %v", err)
	}

	// ── 1. Knowledge base (SNI, origin profiles, market context) ────────────
	chunks := services.KnowledgeChunks
	log.Printf("[knowledge] seeding %d chunks…", len(chunks))
	const batchSize = 5
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]
		ids := make([]string, len(batch))
		for j, ch := range batch {
			ids[j] = ch.ID
		}
		log.Printf("  [knowledge] batch %d/%d: %s",
			i/batchSize+1, (len(chunks)+batchSize-1)/batchSize,
			strings.Join(ids, ", "))
		if err := rag.Upsert(ctx, batch); err != nil {
			log.Fatalf("upsert knowledge batch at %d: %v", i, err)
		}
	}

	// ── 2. Bean sourcing catalog (flavor profiles + supplier contacts) ───────
	log.Println("[beans] ensuring bean catalog collection…")
	if err := rag.EnsureBeanCollection(ctx); err != nil {
		log.Fatalf("ensure bean collection: %v", err)
	}

	beans := services.BeanCatalog
	log.Printf("[beans] seeding %d bean catalog entries…", len(beans))
	for i := 0; i < len(beans); i += batchSize {
		end := i + batchSize
		if end > len(beans) {
			end = len(beans)
		}
		batch := beans[i:end]
		ids := make([]string, len(batch))
		for j, b := range batch {
			ids[j] = b.ID
		}
		log.Printf("  [beans] batch %d/%d: %s",
			i/batchSize+1, (len(beans)+batchSize-1)/batchSize,
			strings.Join(ids, ", "))
		if err := rag.UpsertBeanCatalog(ctx, batch); err != nil {
			log.Fatalf("upsert bean batch at %d: %v", i, err)
		}
	}

	fmt.Printf("\nDone.\n")
	fmt.Printf("  roasty_knowledge : %d chunks\n", len(chunks))
	fmt.Printf("  roasty_beans     : %d entries\n", len(beans))
	fmt.Println("\nStart the API server, then ask the agent:")
	fmt.Println(`  "I need a fruity arabika for pour over around 90k/kg, where can I buy it?"`)
	fmt.Println(`  "What's a cheaper alternative to Toraja with similar body?"`)
	fmt.Println(`  "Recommend beans for a high-volume espresso blend under 50k/kg"`)
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
