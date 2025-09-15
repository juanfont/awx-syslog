package awxsyslog

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type App struct {
	cfg *Config
}

func NewApp(cfg *Config) (*App, error) {
	return &App{cfg: cfg}, nil
}

func (a *App) Serve() error {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.HandleFunc("/", a.handleEndpoint)

	srv := &http.Server{
		Addr:    a.cfg.ListenAddr,
		Handler: router,
	}
	return srv.ListenAndServe()
}

func (a *App) handleEndpoint(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().
		Any("body", body).
		Msg("Received log")
	awxSyslogEventsReceived.Inc()

	loggerType, commonFields, data, err := parseAWXLog(body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse AWX log")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info().Msgf("Parsed log type: %s", loggerType)

	// Convert to syslog message
	syslogMsg := a.awxEventToSyslogMessage(loggerType, data, commonFields)
	msgStr, err := syslogMsg.String()
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize syslog message")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info().Msgf("Generated syslog message: %s", msgStr)

	switch a.cfg.Syslog.Protocol {
	case "tcp":
		conn, err := net.Dial("tcp", a.cfg.Syslog.ServerAddr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to syslog server")
		}
		defer conn.Close()
		conn.Write([]byte(msgStr))
	case "udp":
		conn, err := net.Dial("udp", a.cfg.Syslog.ServerAddr)
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to syslog server")
		}
		defer conn.Close()
		conn.Write([]byte(msgStr))
	case "tls":
		conn, err := tls.Dial("tcp", a.cfg.Syslog.ServerAddr, &tls.Config{})
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect to syslog server")
		}
		defer conn.Close()
		conn.Write([]byte(msgStr))
	default:
		log.Error().Msg("Invalid syslog protocol")
		http.Error(w, "Invalid syslog protocol", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
