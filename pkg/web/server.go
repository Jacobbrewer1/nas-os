package web

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/jacobbrewer1/web/logging"
)

//go:embed templates
var templatesFS embed.FS

func NewServer(l *slog.Logger, addr, port, smbAddr, provider string) *http.Server {
	srv := &http.Server{
		Addr:              addr + ":" + port,
		Handler:           handleIndex(l, smbAddr, provider),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return srv
}

func handleIndex(l *slog.Logger, smbTarget, provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("Handling index request",
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("smb_target", smbTarget),
		)

		type pageData struct {
			SMBTarget string
			Provider  string
		}

		data := &pageData{
			SMBTarget: smbTarget,
			Provider:  provider,
		}

		if err := template.Must(template.New("index.gohtml"). // nolint:revive // Bad lint
									ParseFS(templatesFS, "templates/index.gohtml")).
									Execute(w, data); err != nil {
			l.Error("Error rendering index template", slog.Any(logging.KeyError, err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
