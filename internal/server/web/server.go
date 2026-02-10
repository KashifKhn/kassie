package web

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/KashifKhn/kassie/internal/shared/logger"
)

type ServerConfig struct {
	Host           string
	Port           int
	AllowedOrigins []string
}

type Server struct {
	cfg    *ServerConfig
	server *http.Server
	logger *logger.Logger
}

func NewServer(cfg *ServerConfig, log *logger.Logger) (*Server, error) {
	return &Server{
		cfg:    cfg,
		logger: log,
	}, nil
}

func (s *Server) Start() error {
	distFS, err := GetDistFS()
	if err != nil {
		return fmt.Errorf("failed to get dist filesystem: %w", err)
	}

	if !HasAssets() {
		return fmt.Errorf("no web assets found, run 'make web' first")
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", s.spaHandler(distFS))

	handler := s.corsMiddleware(mux)

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.With().Str("address", addr).Logger().Info("web server listening")

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping web server")

	if s.server == nil {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	s.logger.Info("web server stopped")
	return nil
}

func (s *Server) spaHandler(distFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := strings.TrimPrefix(r.URL.Path, "/")
		if filePath == "" {
			filePath = "index.html"
		}

		file, err := distFS.Open(filePath)
		if err != nil {
			indexFile, indexErr := distFS.Open("index.html")
			if indexErr != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, copyErr := io.Copy(w, indexFile); copyErr != nil {
				s.logger.With().Err(copyErr).Logger().Error("failed to write index.html")
			}
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if stat.IsDir() {
			indexFile, indexErr := distFS.Open(path.Join(filePath, "index.html"))
			if indexErr != nil {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, copyErr := io.Copy(w, indexFile); copyErr != nil {
				s.logger.With().Err(copyErr).Logger().Error("failed to write index.html")
			}
			return
		}

		contentType := mime.TypeByExtension(path.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)

		if strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}

		if _, err := io.Copy(w, file); err != nil {
			s.logger.With().Err(err).Str("file", filePath).Logger().Error("failed to write file")
		}
	}
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if s.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(s.cfg.AllowedOrigins) == 0 || s.cfg.AllowedOrigins[0] == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) isOriginAllowed(origin string) bool {
	if len(s.cfg.AllowedOrigins) == 0 {
		return true
	}

	for _, allowed := range s.cfg.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}
