package webhook

import (
	"context"
	"fmt"
	"log/slog"
	"net"
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
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Try to listen on the configured port with retries.
	// On Windows, a killed process can take time to release its port.
	var listenerStarted bool
	for attempt := 0; attempt < 5; attempt++ {
		if attempt > 0 {
			s.mu.Unlock()
			time.Sleep(time.Duration(attempt) * time.Second)
			s.mu.Lock()
		}

		listener, err := net.Listen("tcp", s.config.ListenAddr())
		if err != nil {
			slog.Warn("failed to listen on port, retrying", "addr", s.config.ListenAddr(), "attempt", attempt+1, "error", err)
			continue
		}

		s.httpServer = &http.Server{
			Addr:         s.config.ListenAddr(),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 3 * time.Second,
		}

		s.running = true
		listenerStarted = true

		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.mu.Lock()
					s.running = false
					s.mu.Unlock()
				}
			}()
			if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				s.mu.Lock()
				s.running = false
				s.mu.Unlock()
			}
		}()
		break
	}

	if !listenerStarted {
		return fmt.Errorf("failed to listen on %s after 5 attempts", s.config.ListenAddr())
	}

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
