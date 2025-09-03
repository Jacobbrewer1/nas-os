package nas

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/caarlos0/env/v10"
	"github.com/gorilla/mux"

	"github.com/jacobbrewer1/nas-os/pkg/storage"
	"github.com/jacobbrewer1/web/logging"
)

type (
	SMBHandlerFunc = func(ctx context.Context, l *slog.Logger, w http.ResponseWriter, r *http.Request)

	// ServerConfig holds configuration for the NAS server
	ServerConfig struct {
		Addr          string `env:"SMB_ADDR" envDefault:"0.0.0.0"`
		Port          int    `env:"SMB_PORT" envDefault:"445"`
		MaxUploadSize int64  `env:"MAX_UPLOAD_SIZE" envDefault:"104857600"` // 100 MB
	}

	// Server represents the NAS server
	Server struct {
		l       *slog.Logger
		cfg     *ServerConfig
		storage storage.Backend
	}
)

// NewSMBServer creates a new instance of the NAS server
func NewSMBServer(
	logger *slog.Logger,
	storageBackend storage.Backend,
) *Server {
	logger = logger.With(
		slog.String("component", "nas_server"),
	)

	cfg := new(ServerConfig)
	if err := env.Parse(cfg); err != nil {
		logger.Error("Failed to parse server configuration from environment",
			slog.Any(logging.KeyError, err),
		)
	}

	return &Server{
		l:       logger,
		cfg:     cfg,
		storage: storageBackend,
	}
}

// SetupRoutes sets up the HTTP routes for the NAS server
func (s *Server) SetupRoutes(r *mux.Router) {
	// Add CORS middleware
	r.Use(corsMiddleware)

	// File operations endpoints
	filesPath := r.Path("/files/{path:.*}")
	filesPath.Methods(http.MethodGet).HandlerFunc(wrapHandler(
		s.l, "download_file", downloadFile(s.storage),
	))
	filesPath.Methods(http.MethodPut).HandlerFunc(wrapHandler(
		s.l, "upload_file", uploadFile(s.storage),
	))
	filesPath.Methods(http.MethodDelete).HandlerFunc(wrapHandler(
		s.l, "delete_file", deleteFile(s.storage),
	))

	r.Methods(http.MethodPost).Path("/upload/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_upload", handleUpload(s.storage, s.cfg.MaxUploadSize),
	))
	r.Methods(http.MethodGet).Path("/download/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_download", downloadFile(s.storage),
	))
	r.Methods(http.MethodGet).Path("/list/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_list", listDirectory(s.storage),
	))
	r.Methods(http.MethodPost).Path("/mkdir/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_mkdir", createDirectory(s.storage),
	))
	r.Methods(http.MethodDelete).Path("/delete/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_delete", handleDelete(s.storage),
	))
	r.Methods(http.MethodGet).Path("/info/{path:.*}").HandlerFunc(wrapHandler(
		s.l, "handle_info", handleGetInfo(s.storage),
	))

	// Handle preflight OPTIONS requests for CORS
	r.Methods(http.MethodOptions).PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// corsMiddleware adds CORS headers to allow cross-origin requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (you may want to restrict this in production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func wrapHandler(l *slog.Logger, name string, fn SMBHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(r.Context(), logging.LoggerWithComponent(l, name), w, r)
	}
}
