package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	ragCollection = "roasty_knowledge"
	ragVectorSize = 1536
	ragTopK       = 3
	embedModel    = "openai/text-embedding-3-small"
)

// KnowledgeChunk is one indexable document fragment stored in Qdrant.
type KnowledgeChunk struct {
	ID     string
	Source string
	Title  string
	Text   string
}

// RAGService wraps Qdrant HTTP REST + OpenRouter embeddings.
// All calls are stateless — the service can be shared across goroutines.
type RAGService struct {
	qdrantURL string
	ai        *AIClient
	http      *http.Client
}

func NewRAGService(qdrantURL string, ai *AIClient) *RAGService {
	return &RAGService{
		qdrantURL: strings.TrimRight(qdrantURL, "/"),
		ai:        ai,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

// embed calls OpenRouter's /embeddings endpoint with text-embedding-3-small.
func (r *RAGService) embed(ctx context.Context, text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]any{
		"model": embedModel,
		"input": text,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.ai.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.ai.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/poisonaifih/roasty")
	req.Header.Set("X-Title", "Roasty")

	res, err := r.ai.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("embed: %s", parsed.Error.Message)
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("embed: empty response")
	}
	return parsed.Data[0].Embedding, nil
}

// qdrantDo sends a JSON request to Qdrant and returns the raw response body.
func (r *RAGService) qdrantDo(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var body *bytes.Reader
	if payload != nil {
		b, _ := json.Marshal(payload)
		body = bytes.NewReader(b)
	} else {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, r.qdrantURL+path, body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return r.http.Do(req)
}

// EnsureCollection creates the Qdrant collection if it does not yet exist.
func (r *RAGService) EnsureCollection(ctx context.Context) error {
	res, err := r.qdrantDo(ctx, http.MethodGet, "/collections/"+ragCollection, nil)
	if err != nil {
		return fmt.Errorf("qdrant check: %w", err)
	}
	res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}

	res, err = r.qdrantDo(ctx, http.MethodPut, "/collections/"+ragCollection, map[string]any{
		"vectors": map[string]any{
			"size":     ragVectorSize,
			"distance": "Cosine",
		},
	})
	if err != nil {
		return fmt.Errorf("qdrant create: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("qdrant create: status %d", res.StatusCode)
	}
	return nil
}

// Upsert embeds every chunk and writes it to Qdrant in one batch call.
// Existing points with the same ID are overwritten (idempotent).
func (r *RAGService) Upsert(ctx context.Context, chunks []KnowledgeChunk) error {
	type point struct {
		ID      uint64         `json:"id"`
		Vector  []float32      `json:"vector"`
		Payload map[string]any `json:"payload"`
	}

	points := make([]point, 0, len(chunks))
	for _, ch := range chunks {
		vec, err := r.embed(ctx, ch.Title+"\n\n"+ch.Text)
		if err != nil {
			return fmt.Errorf("embed %s: %w", ch.ID, err)
		}
		points = append(points, point{
			ID:     fnv64(ch.ID),
			Vector: vec,
			Payload: map[string]any{
				"chunk_id": ch.ID,
				"source":   ch.Source,
				"title":    ch.Title,
				"text":     ch.Text,
			},
		})
	}

	res, err := r.qdrantDo(ctx, http.MethodPut,
		"/collections/"+ragCollection+"/points",
		map[string]any{"points": points})
	if err != nil {
		return fmt.Errorf("qdrant upsert: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("qdrant upsert: status %d", res.StatusCode)
	}
	return nil
}

// Query returns the top-K chunks most relevant to question.
func (r *RAGService) Query(ctx context.Context, question string, topK int) ([]KnowledgeChunk, error) {
	if topK <= 0 {
		topK = ragTopK
	}
	vec, err := r.embed(ctx, question)
	if err != nil {
		return nil, err
	}

	res, err := r.qdrantDo(ctx, http.MethodPost,
		"/collections/"+ragCollection+"/points/search",
		map[string]any{
			"vector":       vec,
			"limit":        topK,
			"with_payload": true,
		})
	if err != nil {
		return nil, fmt.Errorf("qdrant search: %w", err)
	}
	defer res.Body.Close()

	var parsed struct {
		Result []struct {
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	out := make([]KnowledgeChunk, 0, len(parsed.Result))
	for _, hit := range parsed.Result {
		var ch KnowledgeChunk
		if v, ok := hit.Payload["chunk_id"].(string); ok {
			ch.ID = v
		}
		if v, ok := hit.Payload["source"].(string); ok {
			ch.Source = v
		}
		if v, ok := hit.Payload["title"].(string); ok {
			ch.Title = v
		}
		if v, ok := hit.Payload["text"].(string); ok {
			ch.Text = v
		}
		out = append(out, ch)
	}
	return out, nil
}

// QueryText returns retrieved chunks as a single formatted string ready to
// inject into an LLM prompt. Returns empty string on any error so callers
// can fall back to answering without context.
func (r *RAGService) QueryText(ctx context.Context, question string) string {
	chunks, err := r.Query(ctx, question, ragTopK)
	if err != nil || len(chunks) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, ch := range chunks {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		fmt.Fprintf(&sb, "[%s — %s]\n%s", ch.Source, ch.Title, ch.Text)
	}
	return sb.String()
}

// ── Bean catalog collection ──────────────────────────────────────────────────

const beanCollection = "roasty_beans"

// EnsureBeanCollection creates the bean catalog collection in Qdrant if absent.
func (r *RAGService) EnsureBeanCollection(ctx context.Context) error {
	res, err := r.qdrantDo(ctx, http.MethodGet, "/collections/"+beanCollection, nil)
	if err != nil {
		return fmt.Errorf("qdrant check beans: %w", err)
	}
	res.Body.Close()
	if res.StatusCode == http.StatusOK {
		return nil
	}

	res, err = r.qdrantDo(ctx, http.MethodPut, "/collections/"+beanCollection, map[string]any{
		"vectors": map[string]any{
			"size":     ragVectorSize,
			"distance": "Cosine",
		},
	})
	if err != nil {
		return fmt.Errorf("qdrant create beans: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("qdrant create beans: status %d", res.StatusCode)
	}
	return nil
}

// UpsertBeanCatalog embeds every BeanCatalogEntry and writes it to Qdrant.
// The embedding text is the rich narrative document; the full structured entry
// is stored as payload so it can be returned verbatim to the agent.
func (r *RAGService) UpsertBeanCatalog(ctx context.Context, entries []BeanCatalogEntry) error {
	type point struct {
		ID      uint64         `json:"id"`
		Vector  []float32      `json:"vector"`
		Payload map[string]any `json:"payload"`
	}

	points := make([]point, 0, len(entries))
	for _, e := range entries {
		vec, err := r.embed(ctx, e.document())
		if err != nil {
			return fmt.Errorf("embed bean %s: %w", e.ID, err)
		}
		points = append(points, point{
			ID:     fnv64("bean:" + e.ID),
			Vector: vec,
			Payload: map[string]any{
				"entry_id": e.ID,
				"json":     e.resultJSON(),
			},
		})
	}

	res, err := r.qdrantDo(ctx, http.MethodPut,
		"/collections/"+beanCollection+"/points",
		map[string]any{"points": points})
	if err != nil {
		return fmt.Errorf("qdrant upsert beans: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return fmt.Errorf("qdrant upsert beans: status %d", res.StatusCode)
	}
	return nil
}

// FindSimilarBeans returns the top-K bean catalog entries most semantically
// similar to the query. Each result already carries supplier contact details.
func (r *RAGService) FindSimilarBeans(ctx context.Context, query string, topK int) (string, error) {
	if topK <= 0 {
		topK = 4
	}
	vec, err := r.embed(ctx, query)
	if err != nil {
		return "", err
	}

	res, err := r.qdrantDo(ctx, http.MethodPost,
		"/collections/"+beanCollection+"/points/search",
		map[string]any{
			"vector":       vec,
			"limit":        topK,
			"with_payload": true,
		})
	if err != nil {
		return "", fmt.Errorf("qdrant search beans: %w", err)
	}
	defer res.Body.Close()

	var parsed struct {
		Result []struct {
			Score   float32        `json:"score"`
			Payload map[string]any `json:"payload"`
		} `json:"result"`
	}
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Result) == 0 {
		return "[]", nil
	}

	// Collect the pre-serialised JSON entries from payload and wrap in array.
	var parts []string
	for _, hit := range parsed.Result {
		if v, ok := hit.Payload["json"].(string); ok {
			parts = append(parts, v)
		}
	}
	return "[" + strings.Join(parts, ",") + "]", nil
}

// fnv64 returns a deterministic uint64 from a string (FNV-1a).
func fnv64(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}
