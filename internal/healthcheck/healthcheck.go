package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	startTime  time.Time
	port       int
}

type healthResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

func New(port int) *Server {
	if port <= 0 {
		port = 8080
	}

	s := &Server{
		startTime: time.Now(),
		port:      port,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("Health check server starting on port %d", s.port)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health check server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) Stop() error {
	log.Println("Stopping health check server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := healthResponse{
		Status: "healthy",
		Uptime: time.Since(s.startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
