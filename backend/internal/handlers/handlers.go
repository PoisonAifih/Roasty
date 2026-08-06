package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/poisonaifih/roasty/backend/internal/services"
)

type API struct {
	scout *services.ScoutService
	inv   *services.InventoryService
	crm   *services.CRMService
	pool  *pgxpool.Pool
}

func New(scout *services.ScoutService, inv *services.InventoryService, crm *services.CRMService, pool *pgxpool.Pool) *API {
	return &API{scout: scout, inv: inv, crm: crm, pool: pool}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", a.health)
	mux.HandleFunc("GET /api/beans", a.beans)
	mux.HandleFunc("POST /api/scout/recommend", a.recommend)
	mux.HandleFunc("GET /api/scout/shops", a.scoutShops)
	mux.HandleFunc("GET /api/inventory/suggestions", a.inventory)
	mux.HandleFunc("GET /api/crm/follow-ups", a.followUps)
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
