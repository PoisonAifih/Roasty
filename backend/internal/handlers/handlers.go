package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poisonaifih/roasty/backend/internal/services"
)

type API struct {
	scout *services.ScoutService
	inv   *services.InventoryService
	crm   *services.CRMService
	agent *services.Agent
	pool  *pgxpool.Pool
}

func New(scout *services.ScoutService, inv *services.InventoryService, crm *services.CRMService, agent *services.Agent, pool *pgxpool.Pool) *API {
	return &API{scout: scout, inv: inv, crm: crm, agent: agent, pool: pool}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/beans", a.beans)
	mux.HandleFunc("POST /api/scout/recommend", a.recommend)
	mux.HandleFunc("GET /api/scout/shops", a.scoutShops)
	mux.HandleFunc("GET /api/inventory/suggestions", a.inventory)
	mux.HandleFunc("GET /api/crm/follow-ups", a.followUps)

	// Mutating routes: these let the agent act on its recommendations.
	mux.HandleFunc("POST /api/sales", a.recordSale)
	mux.HandleFunc("PATCH /api/beans/{id}/stock", a.adjustStock)
	mux.HandleFunc("POST /api/shops/{id}/contacted", a.markContacted)

	mux.HandleFunc("POST /api/scout/basket", a.basket)

	// Agent: SSE so the UI can render each tool call as it happens.
	mux.HandleFunc("GET /api/agent/stream", a.agentStream)
}

func (a *API) basket(w http.ResponseWriter, r *http.Request) {
	var in services.BasketInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	plans, err := a.scout.BuildBaskets(r.Context(), in)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

// agentStream runs the agent and pushes every trace event to the browser as
// server-sent events. EventSource only issues GET, so the goal arrives as a
// query parameter.
func (a *API) agentStream(w http.ResponseWriter, r *http.Request) {
	goal := r.URL.Query().Get("goal")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// Run invokes emit synchronously, so writing here needs no extra locking.
	send := func(ev services.TraceEvent) {
		payload, err := json.Marshal(ev)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", payload)
		flusher.Flush()
	}

	if _, err := a.agent.Run(r.Context(), goal, send); err != nil {
		send(services.TraceEvent{Type: "error", Message: err.Error()})
	}

	fmt.Fprint(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) beans(w http.ResponseWriter, r *http.Request) {
	beans, err := a.scout.ListBeans(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, beans)
}

func (a *API) recommend(w http.ResponseWriter, r *http.Request) {
	var in services.RecommendInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	recs, err := a.scout.Recommend(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (a *API) scoutShops(w http.ResponseWriter, r *http.Request) {
	origin := r.URL.Query().Get("origin")
	variety := r.URL.Query().Get("variety")
	shops, err := a.scout.FindShops(r.Context(), origin, variety)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, shops)
}

func (a *API) inventory(w http.ResponseWriter, r *http.Request) {
	sugs, err := a.inv.Suggestions(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, sugs)
}

func (a *API) followUps(w http.ResponseWriter, r *http.Request) {
	items, err := a.crm.FollowUps(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (a *API) recordSale(w http.ResponseWriter, r *http.Request) {
	var in services.RecordSaleInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := a.inv.RecordSale(r.Context(), in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

func (a *API) adjustStock(w http.ResponseWriter, r *http.Request) {
	var in services.AdjustStockInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if err := a.inv.AdjustStock(r.Context(), r.PathValue("id"), in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) markContacted(w http.ResponseWriter, r *http.Request) {
	if err := a.crm.MarkContacted(r.Context(), r.PathValue("id")); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "contacted"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
