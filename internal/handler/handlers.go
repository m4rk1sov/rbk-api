package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/m4rk1sov/rbk-api/internal/repository"
	"github.com/m4rk1sov/rbk-api/internal/service"
	"github.com/m4rk1sov/rbk-api/pkg/jsonlog"
	"github.com/m4rk1sov/rbk-api/pkg/util"
	"strconv"
	"time"

	"net/http"
)

type adviceDTO struct {
	Advice string `json:"advice"`
}

type musclesDTO struct {
	Muscles []string `json:"muscles"`
}

type Handler struct {
	r       *chi.Mux
	svc     *service.FitnessService
	logger  *jsonlog.Logger
	started time.Time
}

func New(svc *service.FitnessService, logger *jsonlog.Logger) *Handler {
	h := &Handler{
		r:       chi.NewRouter(),
		svc:     svc,
		logger:  logger,
		started: time.Now(),
	}

	// Middlewares
	h.r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
	h.r.Use(repository.RequestLogger(logger))

	// Routes
	// swagger routes
	h.r.Get("/docs", DocsHandler)
	h.r.Get("/redoc", RedocHandler)
	h.r.Get("/swagger.yaml", SwaggerYAMLHandler)

	h.r.Get("/healthz", h.health)
	h.r.Get("/exercises/{muscle}", h.getExercises)

	// List of available muscles
	h.r.Get("/exercises", h.listMuscles)
	h.r.Get("/exercises/", h.listMuscles)

	// Redirect all unknown routes to /exercises
	h.r.NotFound(h.redirectToExercises)

	h.r.Get("/advice", h.getAdvice)
	return h
}

func (h *Handler) Router() http.Handler {
	return h.r
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	util.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"uptime":  time.Since(h.started).String(),
		"service": "rbk-api",
	})
}

func (h *Handler) getExercises(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	muscle := chi.URLParam(r, "muscle")
	limit := 20
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	resp, err := h.svc.GetExercisesByMuscle(ctx, muscle, limit)
	if err != nil {
		h.logger.PrintError("failed to get exercises", map[string]string{"muscle": muscle})
		util.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	util.WriteJSON(w, http.StatusOK, resp)
}

// GET /exercises or /exercises/
func (h *Handler) listMuscles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	muscles := service.GetAvailableMuscles()

	_ = json.NewEncoder(w).Encode(musclesDTO{Muscles: muscles})
}

// NotFound -> redirect to /exercises
func (h *Handler) redirectToExercises(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		http.Redirect(w, r, "/exercises", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/exercises", http.StatusTemporaryRedirect)
}

func (h *Handler) getAdvice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(adviceDTO{Advice: service.GetAdvice()})
}
