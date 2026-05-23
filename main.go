package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"sentinelmesh/config"
	"sentinelmesh/monitor"
	"sentinelmesh/server"
	"sentinelmesh/store"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

func main() {
	cfg := config.Load()

	db, err := store.New(cfg.DB.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	sse := server.NewSSEManager()

	handler := server.NewHandler(db, sse, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/investigate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.StartInvestigation(w, r)
	})

	mux.HandleFunc("/api/investigations/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/investigations/")
		parts := strings.Split(strings.Trim(path, "/"), "/")

		switch r.Method {
		case http.MethodGet:
			if len(parts) >= 2 && parts[1] == "findings" {
				handler.GetFindings(w, r)
			} else if len(parts) >= 2 && parts[1] == "rules" {
				handler.GetRules(w, r)
			} else if len(parts) >= 2 && parts[1] == "export" {
				handler.ExportReport(w, r)
			} else if len(parts) >= 1 && parts[0] != "" {
				handler.GetInvestigation(w, r)
			} else {
				handler.ListInvestigations(w, r)
			}
		case http.MethodDelete:
			handler.DeleteInvestigation(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/investigations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListInvestigations(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/findings/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.SearchFindings(w, r)
	})

	mux.HandleFunc("/api/alerts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.GetAlerts(w, r)
	})

	mux.HandleFunc("/api/alerts/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ack") {
			if r.Method != http.MethodPut {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler.AcknowledgeAlert(w, r)
		} else {
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler.GetAlerts(w, r)
		}
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		handler.GetStats(w, r)
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		handler.HealthCheck(w, r)
	})

	// Settings routes
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetSettings(w, r)
		case http.MethodPost:
			handler.SaveSettings(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/monitors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.ListMonitors(w, r)
		case http.MethodPost:
			handler.AddMonitor(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/monitors/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handler.RemoveMonitor(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	sched := monitor.NewScheduler(db, handler.RunInvestigationPublic)
	sched.Start()
	defer sched.Stop()

	mux.HandleFunc("/api/nim/models", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.ListNIMModels(w, r)
	})

	mux.HandleFunc("/api/nim/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handler.TestNIMConnection(w, r)
	})

	mux.HandleFunc("/events", handler.SSEHandler)

	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Printf("Warning: frontend not embedded, running API-only mode: %v", err)
	} else {
		fileServer := http.FileServer(http.FS(distFS))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/events" {
				http.NotFound(w, r)
				return
			}
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			if _, err := distFS.Open(strings.TrimPrefix(path, "/")); err != nil {
				r.URL.Path = "/"
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("SentinelMesh starting on http://%s", addr)
	log.Printf("LLM Provider: %s", cfg.LLM.Provider)
	log.Printf("Database: %s", cfg.DB.Path)

	// Apply middleware: RateLimit -> CORS -> Auth
	rateLimiter := server.NewRateLimitMiddleware(cfg.RateLimit)
	handlerChain := rateLimiter.Middleware(server.CORSMiddleware(server.AuthMiddleware(mux)))

	srv := &http.Server{
		Addr:              addr,
		Handler:           handlerChain,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	tlsCert := os.Getenv("TLS_CERT_FILE")
	tlsKey := os.Getenv("TLS_KEY_FILE")
	if tlsCert != "" && tlsKey != "" {
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		log.Printf("SentinelMesh starting on https://%s", addr)
		if err := srv.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		log.Printf("WARNING: No TLS configured. Set TLS_CERT_FILE and TLS_KEY_FILE for production.")
		log.Printf("SentinelMesh starting on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}
