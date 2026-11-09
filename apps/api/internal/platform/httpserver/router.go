package httpserver

import (
	"net/http"
	"os"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/httputil"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"riders.agency/pulse/api/internal/platform/demomode"
	"riders.agency/pulse/api/internal/platform/problem"
)

type HTTPModules interface {
	RegisterHTTP(chi.Router, Dependencies)
}

func NewRouter(deps Dependencies, modules HTTPModules) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://pulse.riders.agency", "http://127.0.0.1:9821", "http://localhost:9821"},
		AllowOriginFunc:  allowPulseOrigin,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "Idempotency-Key", "X-CSRF-Token", "X-Project-Secret", "X-Riders-Timestamp", "X-Riders-Signature", "X-Telegram-Bot-Api-Secret-Token", "X-Relay-Secret"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(demomode.Middleware)
	r.Use(chimiddleware.RequestID)
	r.Use(httputil.TrustedClientIP(os.Getenv("PULSE_TRUSTED_PROXIES")))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(120 * time.Second))

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := chimiddleware.GetReqID(req.Context())
			if requestID == "" {
				requestID = req.Header.Get("X-Request-Id")
			}
			if requestID != "" {
				w.Header().Set("X-Request-Id", requestID)
			}
			w.Header().Set("X-App-Result", "ok")
			w.Header().Set("Content-Language", "ru")
			next.ServeHTTP(w, req)
		})
	})
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		problem.WriteSpec(w, problem.Spec{
			Status: http.StatusNotFound,
			Module: "routing",
			Code:   "route_not_found",
			Title:  "Маршрут не найден",
			Detail: "Для этого адреса нет API-обработчика.",
			Action: "Проверьте путь запроса и активный рабочий контекст.",
			Extra: map[string]any{
				"method": req.Method,
				"path":   req.URL.Path,
			},
		})
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		problem.WriteSpec(w, problem.Spec{
			Status: http.StatusMethodNotAllowed,
			Module: "routing",
			Code:   "method_not_allowed",
			Title:  "Метод не поддерживается",
			Detail: "Маршрут существует, но не принимает этот HTTP-метод.",
			Action: "Используйте доступное действие для этого раздела.",
			Extra: map[string]any{
				"method": req.Method,
				"path":   req.URL.Path,
			},
		})
	})

	mountRoutes(r, deps, modules)
	return r
}

func allowPulseOrigin(r *http.Request, origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}
	switch origin {
	case "https://pulse.riders.agency", "http://127.0.0.1:9821", "http://localhost:9821":
		return true
	}
	path := r.URL.Path
	return path == "/site-chat/widget.js" || strings.HasPrefix(path, "/api/v1/public/site-chat/")
}
