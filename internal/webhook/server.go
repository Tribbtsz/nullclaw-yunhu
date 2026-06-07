package webhook

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/yunhu-channel/yunhu-channel/internal/config"
)

type Server struct {
	config       *config.Config
	runtime      *config.Runtime
	mux          *http.ServeMux
	httpServer   *http.Server
	notifyFn     func([]byte)
	running      bool
	mu           sync.Mutex
}

func NewServer(cfg *config.Config, rt *config.Runtime, notifyFn func([]byte)) *Server {
	return &Server{
		config:   cfg,
		runtime:  rt,
		notifyFn: notifyFn,
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	mux := http.NewServeMux()
	h := NewWebhookHandler(s.config, s.runtime, s.notifyFn)
	mux.HandleFunc(s.config.WebhookPath, h.ServeHTTP)

	s.httpServer = &http.Server{
		Addr:         s.config.ListenAddr(),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	s.running = true

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.mu.Lock()
				s.running = false
				s.mu.Unlock()
			}
		}()
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
