package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type server struct {
	log    *zap.Logger
	port   string
	router *chi.Mux
}

func NewServer(log *zap.Logger, port string) *server {
	return &server{
		log:    log,
		port:   port,
		router: chi.NewRouter(),
	}
}

func (s *server) Start() error {
	s.router.Use(
		middleware.Recoverer,
		middleware.Logger,
	)

	// -------------------------------------------------------------------------
	// App Route

	// This route serves the HTML file in app/ui.html
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// w.Write([]byte("Hello, World!"))
		http.ServeFile(w, r, "app/ui.html")
	})

	// -------------------------------------------------------------------------
	// Worm Positions Route
	s.router.Route("/worm", func(r chi.Router) {
		r.Get("/positions", s.getWormPositions)
	})

	return http.ListenAndServe(":"+s.port, s.router)
}

func (s *server) getWormPositions(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder for the actual implementation
	w.Write([]byte("[]"))
}
