package openldap_exporter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	log "github.com/sirupsen/logrus"
)

var commit string
var tag string

func GetVersion() string {
	return fmt.Sprintf("%s (%s)", tag, commit)
}

type Server struct {
	server  *http.Server
	logger  log.FieldLogger
	cfgPath string
}

func NewMetricsServer(bindAddr, metricsPath, tlsConfigPath string) *Server {
	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.Handler())
	mux.HandleFunc("/version", showVersion)
	return &Server{
		server:  &http.Server{Addr: bindAddr, Handler: mux},
		logger:  log.WithField("component", "server"),
		cfgPath: tlsConfigPath,
	}
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, GetVersion())
}

func (s *Server) Start() error {
	s.logger.WithField("addr", s.server.Addr).Info("starting http listener")
	serverAddress := []string{s.server.Addr}
	useSystemdSocket := false
	var flags web.FlagConfig = web.FlagConfig{
		WebListenAddresses: &serverAddress,
		WebSystemdSocket:   &useSystemdSocket,
		WebConfigFile:      &s.cfgPath,
	}
	err := web.ListenAndServe(s.server, &flags, slog.New(&logrusHandler{s.logger}))
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	s.server.Shutdown(ctx)
	cancel()
}

type logrusHandler struct {
	logger log.FieldLogger
}

func (h *logrusHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *logrusHandler) Handle(_ context.Context, r slog.Record) error {
	fields := log.Fields{}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	ll := h.logger.WithFields(fields)
	switch r.Level {
	case slog.LevelError:
		ll.Error(r.Message)
	case slog.LevelWarn:
		ll.Warn(r.Message)
	case slog.LevelDebug:
		ll.Debug(r.Message)
	default:
		ll.Info(r.Message)
	}
	return nil
}

func (h *logrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := log.Fields{}
	for _, a := range attrs {
		fields[a.Key] = a.Value.Any()
	}
	return &logrusHandler{logger: h.logger.WithFields(fields)}
}

func (h *logrusHandler) WithGroup(_ string) slog.Handler { return h }
