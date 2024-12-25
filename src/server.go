package src

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type server struct {
	log    *zap.Logger
	port   string
	router *chi.Mux
	db     *dbManager
}

func NewServer(log *zap.Logger, port string, db *dbManager) *server {
	return &server{
		log:    log,
		port:   port,
		router: chi.NewRouter(),
		db:     db,
	}
}

func (s *server) Start() error {
	s.router.Use(
		middleware.Recoverer,
		middleware.Logger,
	)

	// -------------------------------------------------------------------------
	// Health Route

	s.router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// -------------------------------------------------------------------------
	// App Route

	// This route serves the HTML file in app/ui.html
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "app/ui.html")
	})

	// -------------------------------------------------------------------------
	// Worm Positions Route
	s.router.Route("/worm", func(r chi.Router) {
		r.Get("/positions", s.positions)
		r.Get("/historical", s.historicalPositions)
	})

	return http.ListenAndServe(":"+s.port, s.router)
}

func (s *server) positions(w http.ResponseWriter, r *http.Request) {
	// Parse the ?id= query parameter from the URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < -1 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	positions, err := s.db.fetchPositions(id)
	if err != nil {
		s.log.Error("failed to fetch positions", zap.Error(err))
		http.Error(w, "failed to fetch positions", http.StatusInternalServerError)
		return
	}

	// Encode the positions as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(positions); err != nil {
		http.Error(w, "failed to encode positions", http.StatusInternalServerError)
		return
	}
}

// historicalPositions returns two fields that contains slices of positions:
//   - recent: contains the 100 most recent positions where the last position is the most recent.
//   - historical: contains a sampled set of 400 positions from the entire history.
func (s *server) historicalPositions(w http.ResponseWriter, r *http.Request) {
	latestPosition, err := s.db.getLatestPosition()
	if err != nil {
		s.log.Error("failed to fetch latest position", zap.Error(err))
		http.Error(w, "failed to fetch latest position", http.StatusInternalServerError)
		return
	}

	last100, err := s.db.fetchPositions(latestPosition.ID - 100)
	if err != nil {
		s.log.Error("failed to fetch recent positions", zap.Error(err))
		http.Error(w, "failed to fetch recent positions", http.StatusInternalServerError)
		return
	}

	historical, err := s.db.fetchSample(400)
	if err != nil {
		s.log.Error("failed to fetch historical positions", zap.Error(err))
		http.Error(w, "failed to fetch historical positions", http.StatusInternalServerError)
		return
	}

	type resp struct {
		Recent     []position `json:"recent"`
		Historical []position `json:"historical"`
	}

	// Encode the positions as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp{Recent: last100, Historical: historical}); err != nil {
		http.Error(w, "failed to encode positions", http.StatusInternalServerError)
		return
	}

}
